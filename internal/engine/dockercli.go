package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type DockerCLI struct {
	Bin string
}

func NewDockerCLI() DockerCLI {
	return DockerCLI{Bin: "docker"}
}

func (d DockerCLI) Check(ctx context.Context) RuntimeStatus {
	status := RuntimeStatus{}
	if output, err := exec.CommandContext(ctx, d.Bin, "--version").CombinedOutput(); err == nil {
		status.DockerInstalled = true
		status.DockerVersion = strings.TrimSpace(string(output))
	}
	if status.DockerInstalled {
		if output, err := exec.CommandContext(ctx, d.Bin, "info", "--format", "{{json .ServerVersion}}").CombinedOutput(); err == nil {
			status.EngineRunning = true
			status.DockerVersion = firstNonEmpty(status.DockerVersion, strings.Trim(strings.TrimSpace(string(output)), `"`))
		}
		if output, err := exec.CommandContext(ctx, d.Bin, "compose", "version", "--short").CombinedOutput(); err == nil {
			status.ComposePresent = true
			status.ComposeVersion = strings.TrimSpace(string(output))
		}
	}
	if output, err := exec.CommandContext(ctx, "podman", "--version").CombinedOutput(); err == nil {
		status.PodmanPresent = true
		status.PodmanVersion = strings.TrimSpace(string(output))
	}
	return status
}

func (d DockerCLI) ImageExists(ctx context.Context, image string) (bool, error) {
	_, err := d.run(ctx, nil, "image", "inspect", image)
	if err == nil {
		return true, nil
	}
	return false, nil
}

func (d DockerCLI) Pull(ctx context.Context, image string, output io.Writer) error {
	cmd := exec.CommandContext(ctx, d.Bin, "pull", image)
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Run(); err != nil {
		return explainDockerError("pull image "+image, err, nil)
	}
	return nil
}

func (d DockerCLI) NetworkExists(ctx context.Context, name string) (bool, error) {
	_, err := d.run(ctx, nil, "network", "inspect", name)
	return err == nil, nil
}

func (d DockerCLI) CreateNetwork(ctx context.Context, name string) error {
	_, err := d.run(ctx, nil, "network", "create", name)
	return err
}

func (d DockerCLI) VolumeExists(ctx context.Context, name string) (bool, error) {
	_, err := d.run(ctx, nil, "volume", "inspect", name)
	return err == nil, nil
}

func (d DockerCLI) CreateVolume(ctx context.Context, name string) error {
	_, err := d.run(ctx, nil, "volume", "create", name)
	return err
}

func (d DockerCLI) ContainerExists(ctx context.Context, name string) (bool, error) {
	_, err := d.run(ctx, nil, "container", "inspect", name)
	return err == nil, nil
}

func (d DockerCLI) Run(ctx context.Context, opts RunOptions) error {
	args := []string{"run", "-d", "--name", opts.Name}
	if opts.Network != "" {
		args = append(args, "--network", opts.Network)
	}
	if opts.Restart != "" {
		args = append(args, "--restart", opts.Restart)
	}
	for containerPort, hostPort := range opts.Ports {
		args = append(args, "-p", fmt.Sprintf("%d:%d", hostPort, containerPort))
	}
	for volume, path := range opts.Volumes {
		args = append(args, "-v", volume+":"+path)
	}
	for key, value := range opts.Environment {
		args = append(args, "-e", key+"="+value)
	}
	for key, value := range opts.Labels {
		args = append(args, "--label", key+"="+value)
	}
	args = append(args, opts.Image)
	args = append(args, opts.Command...)
	_, err := d.run(ctx, nil, args...)
	return err
}

func (d DockerCLI) Start(ctx context.Context, name string) error {
	_, err := d.run(ctx, nil, "start", name)
	return err
}

func (d DockerCLI) Stop(ctx context.Context, name string) error {
	_, err := d.run(ctx, nil, "stop", name)
	return err
}

func (d DockerCLI) Restart(ctx context.Context, name string) error {
	_, err := d.run(ctx, nil, "restart", name)
	return err
}

func (d DockerCLI) Remove(ctx context.Context, name string, volumes bool) error {
	args := []string{"rm", "-f"}
	if volumes {
		args = append(args, "-v")
	}
	args = append(args, name)
	_, err := d.run(ctx, nil, args...)
	return err
}

func (d DockerCLI) Logs(ctx context.Context, name string, follow bool, output io.Writer) error {
	args := []string{"logs", "--tail", "200"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, name)
	cmd := exec.CommandContext(ctx, d.Bin, args...)
	cmd.Stdout = output
	cmd.Stderr = output
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return explainDockerError("read logs for "+name, err, nil)
	}
	return nil
}

func (d DockerCLI) InspectHealth(ctx context.Context, name string) (ContainerHealth, error) {
	format := "{{json .}}"
	output, err := d.run(ctx, nil, "inspect", "--format", format, name)
	if err != nil {
		return ContainerHealth{}, err
	}
	var raw struct {
		ID     string
		Name   string
		Config struct{ Image string }
		State  struct {
			Status    string
			Health    *struct{ Status string }
			StartedAt string
		}
		NetworkSettings struct {
			Ports    map[string][]struct{ HostIP, HostPort string }
			Networks map[string]any
		}
	}
	if err := json.Unmarshal(output, &raw); err != nil {
		return ContainerHealth{}, err
	}
	health := "unavailable"
	if raw.State.Health != nil {
		health = raw.State.Health.Status
	}
	networks := make([]string, 0, len(raw.NetworkSettings.Networks))
	for network := range raw.NetworkSettings.Networks {
		networks = append(networks, network)
	}
	return ContainerHealth{ID: shortID(raw.ID), Name: strings.TrimPrefix(raw.Name, "/"), Image: raw.Config.Image, State: raw.State.Status, Health: health, Ports: formatPorts(raw.NetworkSettings.Ports), Network: strings.Join(networks, ","), StartedAt: raw.State.StartedAt}, nil
}

func (d DockerCLI) Stats(ctx context.Context, names []string) ([]ContainerStats, error) {
	args := []string{"stats", "--no-stream", "--format", "{{json .}}"}
	args = append(args, names...)
	output, err := d.run(ctx, nil, args...)
	if err != nil {
		return nil, err
	}
	var stats []ContainerStats
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var row ContainerStats
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, err
		}
		stats = append(stats, row)
	}
	return stats, nil
}

func (d DockerCLI) List(ctx context.Context, label string) ([]ContainerInfo, error) {
	args := []string{"ps", "-a", "--format", "{{json .}}"}
	if label != "" {
		args = append(args, "--filter", "label="+label)
	}
	output, err := d.run(ctx, nil, args...)
	if err != nil {
		return nil, err
	}
	var containers []ContainerInfo
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var row ContainerInfo
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, err
		}
		containers = append(containers, row)
	}
	return containers, nil
}

func (d DockerCLI) ExecInteractive(ctx context.Context, name string, args []string) error {
	cmdArgs := append([]string{"exec", "-it", name}, args...)
	cmd := exec.CommandContext(ctx, d.Bin, cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (d DockerCLI) run(ctx context.Context, stdin io.Reader, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, d.Bin, args...)
	cmd.Stdin = stdin
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return output, explainDockerError(strings.Join(args, " "), err, stderr.Bytes())
	}
	return output, nil
}

func explainDockerError(action string, err error, stderr []byte) error {
	detail := strings.TrimSpace(string(stderr))
	if detail == "" {
		detail = err.Error()
	}
	var exit *exec.ExitError
	switch {
	case errors.Is(err, exec.ErrNotFound):
		return fmt.Errorf("could not %s: Docker CLI is not installed\nSolution: install Docker Desktop or Docker Engine and ensure docker is on PATH", action)
	case errors.As(err, &exit):
		return fmt.Errorf("could not %s: %s\nSolution: run devstack doctor, confirm Docker Engine is running, and retry", action, detail)
	default:
		return fmt.Errorf("could not %s: %s", action, detail)
	}
}

func formatPorts(ports map[string][]struct{ HostIP, HostPort string }) string {
	var parts []string
	for container, bindings := range ports {
		for _, binding := range bindings {
			parts = append(parts, binding.HostPort+"->"+container)
		}
	}
	return strings.Join(parts, ", ")
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func ClientTool(serviceID string) []string {
	switch serviceID {
	case "postgres":
		return []string{"psql", "-U", "postgres"}
	case "mongo":
		return []string{"mongosh"}
	case "redis":
		return []string{"redis-cli"}
	case "mysql", "mariadb":
		return []string{"mysql", "-u", "root", "-p"}
	default:
		if runtime.GOOS == "windows" {
			return []string{"cmd"}
		}
		return []string{"sh"}
	}
}
