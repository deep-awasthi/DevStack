package app

import (
	"context"
	"fmt"
	"io"
	"sort"

	appconfig "github.com/deepawasthi/devstack/internal/config"
	"github.com/deepawasthi/devstack/internal/engine"
	"github.com/deepawasthi/devstack/internal/services"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	Runtime engine.Runtime
	Catalog services.Catalog
	Store   appconfig.Store
	Log     *logrus.Logger
}

func NewManager(runtime engine.Runtime, catalog services.Catalog, store appconfig.Store, log *logrus.Logger) Manager {
	if log == nil {
		log = logrus.New()
	}
	return Manager{Runtime: runtime, Catalog: catalog, Store: store, Log: log}
}

func (m Manager) Init(ctx context.Context, name string, serviceIDs []string, versions map[string]string) (appconfig.Environment, error) {
	if name == "" {
		name = "default"
	}
	env := appconfig.NewEnvironment(name)
	for _, id := range serviceIDs {
		service, ok := m.Catalog.Get(id)
		if !ok {
			return env, fmt.Errorf("unsupported service %q\nSolution: run devstack search %s to find supported services", id, id)
		}
		version := versions[id]
		if version == "" {
			version = service.DefaultVersion
		}
		env.Services[service.ID] = appconfig.ServiceConfig{Version: version, Enabled: true}
	}
	resolved, env, err := services.ResolveEnvironment(env, m.Catalog)
	if err != nil {
		return env, err
	}
	_ = resolved
	return env, m.Store.Save(env)
}

func (m Manager) LoadResolved() (appconfig.Environment, []services.ResolvedService, error) {
	env, err := m.Store.Load()
	if err != nil {
		return env, nil, err
	}
	resolved, updated, err := services.ResolveEnvironment(env, m.Catalog)
	if err != nil {
		return env, nil, err
	}
	if err := m.Store.Save(updated); err != nil {
		return env, nil, err
	}
	return updated, resolved, nil
}

func (m Manager) Up(ctx context.Context, output io.Writer) error {
	env, resolved, err := m.LoadResolved()
	if err != nil {
		return err
	}
	if len(resolved) == 0 {
		return fmt.Errorf("no enabled services in %s\nSolution: run devstack init --service postgres --service redis", m.Store.Path)
	}
	if err := m.ensureDocker(ctx); err != nil {
		return err
	}
	ok, err := m.Runtime.NetworkExists(ctx, env.Network)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintf(output, "Creating network %s\n", env.Network)
		if err := m.Runtime.CreateNetwork(ctx, env.Network); err != nil {
			return err
		}
	}
	for _, service := range resolved {
		if err := m.ensureService(ctx, service, output); err != nil {
			return err
		}
	}
	return nil
}

func (m Manager) ensureService(ctx context.Context, resolved services.ResolvedService, output io.Writer) error {
	exists, err := m.Runtime.ImageExists(ctx, resolved.Image)
	if err != nil {
		return err
	}
	if !exists {
		fmt.Fprintf(output, "Pulling %s\n", resolved.Image)
		if err := m.Runtime.Pull(ctx, resolved.Image, output); err != nil {
			return err
		}
	} else {
		fmt.Fprintf(output, "Image present: %s\n", resolved.Image)
	}
	for _, volume := range resolved.Service.Volumes {
		name := resolved.VolumeNames[volume.Name]
		exists, err := m.Runtime.VolumeExists(ctx, name)
		if err != nil {
			return err
		}
		if !exists {
			fmt.Fprintf(output, "Creating volume %s\n", name)
			if err := m.Runtime.CreateVolume(ctx, name); err != nil {
				return err
			}
		}
	}
	exists, err = m.Runtime.ContainerExists(ctx, resolved.Container)
	if err != nil {
		return err
	}
	if exists {
		fmt.Fprintf(output, "Starting existing container %s\n", resolved.Container)
		return m.Runtime.Start(ctx, resolved.Container)
	}
	hostPorts := map[int]int{}
	for _, port := range resolved.Service.Ports {
		hostPorts[port.Internal] = resolved.Ports[port.Name]
	}
	volumes := map[string]string{}
	for _, volume := range resolved.Service.Volumes {
		volumes[resolved.VolumeNames[volume.Name]] = volume.Path
	}
	fmt.Fprintf(output, "Creating container %s\n", resolved.Container)
	return m.Runtime.Run(ctx, engine.RunOptions{
		Name:        resolved.Container,
		Image:       resolved.Image,
		Network:     resolved.Network,
		Ports:       hostPorts,
		Volumes:     volumes,
		Environment: services.Environment(resolved),
		Labels: map[string]string{
			"devstack.managed":     "true",
			"devstack.environment": resolved.Network,
			"devstack.service":     resolved.Service.ID,
		},
		Restart: "unless-stopped",
		Command: services.Command(resolved),
	})
}

func (m Manager) Stop(ctx context.Context) error {
	_, resolved, err := m.LoadResolved()
	if err != nil {
		return err
	}
	for _, service := range resolved {
		if err := m.Runtime.Stop(ctx, service.Container); err != nil {
			return err
		}
	}
	return nil
}

func (m Manager) Restart(ctx context.Context) error {
	_, resolved, err := m.LoadResolved()
	if err != nil {
		return err
	}
	for _, service := range resolved {
		if err := m.Runtime.Restart(ctx, service.Container); err != nil {
			return err
		}
	}
	return nil
}

func (m Manager) Remove(ctx context.Context, volumes bool) error {
	_, resolved, err := m.LoadResolved()
	if err != nil {
		return err
	}
	for _, service := range resolved {
		if err := m.Runtime.Remove(ctx, service.Container, volumes); err != nil {
			return err
		}
	}
	return nil
}

func (m Manager) Health(ctx context.Context) ([]engine.ContainerHealth, error) {
	_, resolved, err := m.LoadResolved()
	if err != nil {
		return nil, err
	}
	var rows []engine.ContainerHealth
	for _, service := range resolved {
		row, err := m.Runtime.InspectHealth(ctx, service.Container)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (m Manager) Stats(ctx context.Context) ([]engine.ContainerStats, error) {
	_, resolved, err := m.LoadResolved()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(resolved))
	for _, service := range resolved {
		names = append(names, service.Container)
	}
	return m.Runtime.Stats(ctx, names)
}

func (m Manager) Logs(ctx context.Context, serviceID string, follow bool, output io.Writer) error {
	_, resolved, err := m.LoadResolved()
	if err != nil {
		return err
	}
	for _, service := range resolved {
		if service.Service.ID == serviceID || service.Container == serviceID {
			return m.Runtime.Logs(ctx, service.Container, follow, output)
		}
	}
	return fmt.Errorf("service %q is not enabled in this environment", serviceID)
}

func (m Manager) Connect(ctx context.Context, serviceID string) error {
	_, resolved, err := m.LoadResolved()
	if err != nil {
		return err
	}
	for _, service := range resolved {
		if service.Service.ID == serviceID || service.Container == serviceID {
			return m.Runtime.ExecInteractive(ctx, service.Container, engine.ClientTool(service.Service.ID))
		}
	}
	return fmt.Errorf("service %q is not enabled in this environment", serviceID)
}

func (m Manager) Connections() ([]ConnectionDetails, error) {
	_, resolved, err := m.LoadResolved()
	if err != nil {
		return nil, err
	}
	var details []ConnectionDetails
	for _, service := range resolved {
		item := ConnectionDetails{Service: service.Service.Name, Values: map[string]string{}}
		for _, hint := range service.Service.ConnectionHints {
			item.Values[hint.Label] = services.Render(hint.Pattern, service)
		}
		if len(item.Values) == 0 {
			for _, port := range service.Service.Ports {
				item.Values[port.Name] = fmt.Sprintf("localhost:%d", service.Ports[port.Name])
			}
		}
		details = append(details, item)
	}
	sort.SliceStable(details, func(i, j int) bool { return details[i].Service < details[j].Service })
	return details, nil
}

func (m Manager) ensureDocker(ctx context.Context) error {
	status := m.Runtime.Check(ctx)
	if !status.DockerInstalled {
		return fmt.Errorf("Docker CLI is not installed\nSolution: install Docker Desktop or Docker Engine and ensure docker is on PATH")
	}
	if !status.EngineRunning {
		return fmt.Errorf("Docker Engine is not running\nSolution: start Docker Desktop or the Docker daemon, then run devstack doctor")
	}
	return nil
}

type ConnectionDetails struct {
	Service string
	Values  map[string]string
}
