package engine

import (
	"context"
	"io"
)

type RuntimeStatus struct {
	DockerInstalled bool
	EngineRunning   bool
	ComposePresent  bool
	PodmanPresent   bool
	DockerVersion   string
	ComposeVersion  string
	PodmanVersion   string
}

type ContainerInfo struct {
	ID         string
	Name       string
	Names      string
	Image      string
	State      string
	Status     string
	Ports      string
	CreatedAt  string
	RunningFor string
}

func (c ContainerInfo) DisplayName() string {
	if c.Names != "" {
		return c.Names
	}
	return c.Name
}

type ContainerHealth struct {
	ID        string
	Name      string
	Image     string
	State     string
	Health    string
	Status    string
	Ports     string
	Network   string
	Uptime    string
	StartedAt string
}

type ContainerStats struct {
	Name      string
	CPUPerc   string
	MemUsage  string
	MemPerc   string
	NetIO     string
	BlockIO   string
	PIDs      string
	Container string
}

type RunOptions struct {
	Name        string
	Image       string
	Network     string
	Ports       map[int]int
	Volumes     map[string]string
	Environment map[string]string
	Labels      map[string]string
	Restart     string
	Command     []string
	DependsOn   []string
}

type Runtime interface {
	Check(ctx context.Context) RuntimeStatus
	ImageExists(ctx context.Context, image string) (bool, error)
	Pull(ctx context.Context, image string, output io.Writer) error
	NetworkExists(ctx context.Context, name string) (bool, error)
	CreateNetwork(ctx context.Context, name string) error
	VolumeExists(ctx context.Context, name string) (bool, error)
	CreateVolume(ctx context.Context, name string) error
	ContainerExists(ctx context.Context, name string) (bool, error)
	Run(ctx context.Context, opts RunOptions) error
	Start(ctx context.Context, name string) error
	Stop(ctx context.Context, name string) error
	Restart(ctx context.Context, name string) error
	Remove(ctx context.Context, name string, volumes bool) error
	Logs(ctx context.Context, name string, follow bool, output io.Writer) error
	InspectHealth(ctx context.Context, name string) (ContainerHealth, error)
	Stats(ctx context.Context, names []string) ([]ContainerStats, error)
	List(ctx context.Context, label string) ([]ContainerInfo, error)
	ExecInteractive(ctx context.Context, name string, args []string) error
}
