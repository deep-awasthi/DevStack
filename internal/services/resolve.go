package services

import (
	"fmt"
	"strings"

	appconfig "github.com/deepawasthi/devstack/internal/config"
	"github.com/deepawasthi/devstack/internal/ports"
	"github.com/deepawasthi/devstack/internal/utils"
)

func ResolveEnvironment(env appconfig.Environment, catalog Catalog) ([]ResolvedService, appconfig.Environment, error) {
	allocator := ports.NewAllocator()
	for _, cfg := range env.Services {
		for _, port := range cfg.Ports {
			allocator.Mark(port)
		}
	}
	var resolved []ResolvedService
	for id, cfg := range env.Services {
		if !cfg.Enabled {
			continue
		}
		service, ok := catalog.Get(id)
		if !ok {
			return nil, env, fmt.Errorf("unsupported service %q\nSolution: run devstack search %s to find the nearest supported service", id, id)
		}
		if cfg.Version == "" {
			cfg.Version = service.DefaultVersion
		}
		if !service.SupportsVersion(cfg.Version) {
			return nil, env, fmt.Errorf("%s does not support version %q\nSolution: run devstack inspect %s to see available versions", service.Name, cfg.Version, id)
		}
		if cfg.Ports == nil {
			cfg.Ports = map[string]int{}
		}
		for _, port := range service.Ports {
			if cfg.Ports[port.Name] == 0 {
				allocated, err := allocator.Reserve(port.Preferred)
				if err != nil {
					return nil, env, err
				}
				cfg.Ports[port.Name] = allocated
			}
		}
		cfg.Credentials = ensureCredentials(service, cfg.Credentials)
		env.Services[id] = cfg
		resolved = append(resolved, ResolvedService{
			Service:      service,
			Version:      cfg.Version,
			Image:        service.ImageFor(cfg.Version),
			Container:    "devstack-" + utils.SanitizeName(env.Name) + "-" + service.ID,
			Ports:        cfg.Ports,
			Credentials:  cfg.Credentials,
			Network:      env.Network,
			VolumeNames:  volumeNames(env.Name, service),
			Dependencies: service.DependsOn,
		})
	}
	return resolved, env, nil
}

func ensureCredentials(service Service, existing map[string]string) map[string]string {
	if existing == nil {
		existing = map[string]string{}
	}
	for key, value := range service.Credentials.Defaults {
		if existing[key] == "" {
			existing[key] = value
		}
	}
	for _, key := range credentialKeys(service) {
		if existing[key] == "" {
			existing[key] = utils.RandomSecret(24)
		}
	}
	return existing
}

func credentialKeys(service Service) []string {
	keys := map[string]bool{}
	for key := range service.Environment {
		for _, candidate := range placeholders(key + service.Environment[key]) {
			if strings.Contains(candidate, "password") || strings.Contains(candidate, "secret") {
				keys[candidate] = true
			}
		}
	}
	for _, arg := range service.Command {
		for _, candidate := range placeholders(arg) {
			if strings.Contains(candidate, "password") || strings.Contains(candidate, "secret") {
				keys[candidate] = true
			}
		}
	}
	for _, hint := range service.ConnectionHints {
		for _, candidate := range placeholders(hint.Pattern) {
			if strings.Contains(candidate, "password") || strings.Contains(candidate, "secret") {
				keys[candidate] = true
			}
		}
	}
	if len(keys) == 0 && service.Credentials.Defaults != nil {
		keys["password"] = true
	}
	out := make([]string, 0, len(keys))
	for key := range keys {
		out = append(out, key)
	}
	return out
}

func placeholders(value string) []string {
	var found []string
	for {
		start := strings.Index(value, "{{")
		if start < 0 {
			return found
		}
		end := strings.Index(value[start+2:], "}}")
		if end < 0 {
			return found
		}
		token := value[start+2 : start+2+end]
		if !strings.HasPrefix(token, "port:") {
			found = append(found, token)
		}
		value = value[start+2+end+2:]
	}
}

func volumeNames(envName string, service Service) map[string]string {
	names := map[string]string{}
	prefix := "devstack_" + utils.SanitizeName(envName) + "_" + service.ID
	for _, volume := range service.Volumes {
		names[volume.Name] = prefix + "_" + utils.SanitizeName(volume.Name)
	}
	return names
}

func Render(value string, resolved ResolvedService) string {
	out := value
	for key, credential := range resolved.Credentials {
		out = strings.ReplaceAll(out, "{{"+key+"}}", credential)
	}
	for key, port := range resolved.Ports {
		out = strings.ReplaceAll(out, "{{port:"+key+"}}", fmt.Sprintf("%d", port))
	}
	return out
}

func Environment(resolved ResolvedService) map[string]string {
	env := map[string]string{}
	for key, value := range resolved.Service.Environment {
		env[key] = Render(value, resolved)
	}
	return env
}

func Command(resolved ResolvedService) []string {
	args := make([]string, len(resolved.Service.Command))
	for i, arg := range resolved.Service.Command {
		args[i] = Render(arg, resolved)
	}
	return args
}
