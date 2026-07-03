package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/deepawasthi/devstack/internal/app"
	appconfig "github.com/deepawasthi/devstack/internal/config"
	"github.com/deepawasthi/devstack/internal/engine"
	"github.com/deepawasthi/devstack/internal/generator"
	"github.com/deepawasthi/devstack/internal/services"
	"github.com/deepawasthi/devstack/internal/templates"
	"github.com/deepawasthi/devstack/internal/terminal"
	"github.com/deepawasthi/devstack/internal/ui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "0.1.0"
	cfgFile string
	log     = logrus.New()
)

func Execute() error {
	root := newRootCommand()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:          "devstack",
		Short:        "Terminal-based Docker environment manager",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			catalog := services.NewCatalog()
			runtime := engine.NewDockerCLI()
			result, err := ui.RunSelector(catalog, runtime.Check(ctx))
			if err != nil {
				return err
			}
			if len(result.Services) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No services selected.")
				return nil
			}
			manager := newManager()
			env, err := manager.Init(ctx, firstNonEmpty(result.Name, "dev"), result.Services, map[string]string{})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s with %d service(s). Run devstack up to start.\n", appconfig.FileName, len(env.Services))
			return nil
		},
	}
	root.PersistentFlags().StringVarP(&cfgFile, "config", "c", appconfig.FileName, "path to devstack configuration")
	root.PersistentFlags().String("log-level", "warn", "log level")
	_ = viper.BindPFlag("log-level", root.PersistentFlags().Lookup("log-level"))
	cobra.OnInitialize(func() {
		level, err := logrus.ParseLevel(viper.GetString("log-level"))
		if err == nil {
			log.SetLevel(level)
		}
	})
	root.AddCommand(
		initCommand(),
		upCommand(),
		simpleLifecycleCommand("down", "Stop all services", func(ctx context.Context, m app.Manager) error { return m.Stop(ctx) }),
		simpleLifecycleCommand("stop", "Stop all services", func(ctx context.Context, m app.Manager) error { return m.Stop(ctx) }),
		simpleLifecycleCommand("restart", "Restart all services", func(ctx context.Context, m app.Manager) error { return m.Restart(ctx) }),
		removeCommand(),
		logsCommand(),
		connectCommand(),
		inspectCommand(),
		statsCommand(),
		listCommand(),
		cleanupCommand(),
		updateCommand(),
		exportCommand(),
		doctorCommand(),
		versionCommand(),
		templatesCommand(),
		searchCommand(),
	)
	return root
}

func newManager() app.Manager {
	return app.NewManager(engine.NewDockerCLI(), services.NewCatalog(), appconfig.NewStore(cfgFile), log)
}

func initCommand() *cobra.Command {
	var name string
	var serviceFlags []string
	var templateID string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create devstack.yml",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			serviceIDs := append([]string{}, serviceFlags...)
			if templateID != "" {
				template, ok := templates.Get(templateID)
				if !ok {
					return fmt.Errorf("unknown template %q\nSolution: run devstack templates", templateID)
				}
				serviceIDs = append(serviceIDs, template.Services...)
				if name == "" {
					name = template.ID
				}
			}
			if len(args) > 0 {
				serviceIDs = append(serviceIDs, args...)
			}
			if len(serviceIDs) == 0 {
				return fmt.Errorf("no services selected\nSolution: pass --service postgres or --template developer-essentials")
			}
			env, err := newManager().Init(ctx, firstNonEmpty(name, "dev"), dedupe(serviceIDs), map[string]string{})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s for %q with %d service(s).\n", cfgFile, env.Name, len(env.Services))
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "environment name")
	cmd.Flags().StringArrayVarP(&serviceFlags, "service", "s", nil, "service to enable")
	cmd.Flags().StringVarP(&templateID, "template", "t", "", "template ID to use")
	return cmd
}

func upCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Create and start the environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			return newManager().Up(cmd.Context(), cmd.OutOrStdout())
		},
	}
}

func simpleLifecycleCommand(use, short string, fn func(context.Context, app.Manager) error) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := fn(cmd.Context(), newManager()); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s complete.\n", use)
			return nil
		},
	}
}

func removeCommand() *cobra.Command {
	var volumes bool
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove managed containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := newManager().Remove(cmd.Context(), volumes); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Removed managed containers.")
			return nil
		},
	}
	cmd.Flags().BoolVarP(&volumes, "volumes", "v", false, "remove named volumes too")
	return cmd
}

func logsCommand() *cobra.Command {
	var follow bool
	cmd := &cobra.Command{
		Use:   "logs SERVICE",
		Short: "Show service logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return newManager().Logs(cmd.Context(), args[0], follow, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow logs")
	return cmd
}

func connectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "connect SERVICE",
		Short: "Open a service client inside the container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return newManager().Connect(cmd.Context(), args[0])
		},
	}
}

func inspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect",
		Short: "Show health and connection details",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := newManager()
			rows, err := manager.Health(cmd.Context())
			if err != nil {
				return err
			}
			var tableRows [][]string
			for _, row := range rows {
				tableRows = append(tableRows, []string{row.Name, row.Image, row.State, row.Health, row.Ports, row.Network})
			}
			terminal.Table(cmd.OutOrStdout(), []string{"Name", "Image", "State", "Health", "Ports", "Network"}, tableRows)
			details, err := manager.Connections()
			if err != nil {
				return err
			}
			for _, detail := range details {
				fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", detail.Service)
				keys := make([]string, 0, len(detail.Values))
				for key := range detail.Values {
					keys = append(keys, key)
				}
				for _, key := range keys {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s\n", key, detail.Values[key])
				}
			}
			return nil
		},
	}
}

func statsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show resource usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			stats, err := newManager().Stats(cmd.Context())
			if err != nil {
				return err
			}
			var rows [][]string
			for _, stat := range stats {
				rows = append(rows, []string{stat.Name, stat.CPUPerc, stat.MemUsage, stat.MemPerc, stat.NetIO, stat.BlockIO})
			}
			terminal.Table(cmd.OutOrStdout(), []string{"Name", "CPU", "Memory", "Mem %", "Net IO", "Block IO"}, rows)
			return nil
		},
	}
}

func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List DevStack containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			containers, err := engine.NewDockerCLI().List(cmd.Context(), "devstack.managed=true")
			if err != nil {
				return err
			}
			var rows [][]string
			for _, c := range containers {
				rows = append(rows, []string{c.DisplayName(), c.Image, c.State, c.Status, c.Ports})
			}
			terminal.Table(cmd.OutOrStdout(), []string{"Name", "Image", "State", "Status", "Ports"}, rows)
			return nil
		},
	}
}

func cleanupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "Remove stopped DevStack containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			containers, err := engine.NewDockerCLI().List(cmd.Context(), "devstack.managed=true")
			if err != nil {
				return err
			}
			rt := engine.NewDockerCLI()
			for _, c := range containers {
				if strings.EqualFold(c.State, "exited") || strings.EqualFold(c.State, "created") {
					if err := rt.Remove(cmd.Context(), c.DisplayName(), false); err != nil {
						return err
					}
				}
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Cleanup complete.")
			return nil
		},
	}
}

func updateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Pull configured service images",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, resolved, err := newManager().LoadResolved()
			if err != nil {
				return err
			}
			rt := engine.NewDockerCLI()
			for _, service := range resolved {
				fmt.Fprintf(cmd.OutOrStdout(), "Updating %s\n", service.Image)
				if err := rt.Pull(cmd.Context(), service.Image, cmd.OutOrStdout()); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func exportCommand() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export compose and application configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := newManager()
			env, resolved, err := manager.LoadResolved()
			if err != nil {
				return err
			}
			if err := generator.NewExporter(dir).Export(env, resolved); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Exported Docker Compose and app configuration to %s\n", dir)
			return nil
		},
	}
	cmd.Flags().StringVarP(&dir, "dir", "d", ".", "export directory")
	return cmd
}

func doctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check Docker, Compose, and Podman availability",
		RunE: func(cmd *cobra.Command, args []string) error {
			status := engine.NewDockerCLI().Check(cmd.Context())
			rows := [][]string{
				{"Docker CLI", boolText(status.DockerInstalled), status.DockerVersion},
				{"Docker Engine", boolText(status.EngineRunning), ""},
				{"Docker Compose", boolText(status.ComposePresent), status.ComposeVersion},
				{"Podman", boolText(status.PodmanPresent), status.PodmanVersion},
			}
			terminal.Table(cmd.OutOrStdout(), []string{"Check", "Status", "Version"}, rows)
			if !status.DockerInstalled || !status.EngineRunning {
				return fmt.Errorf("Docker is not ready\nSolution: install/start Docker, then run devstack doctor again")
			}
			return nil
		},
	}
}

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print DevStack version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "devstack %s\n", version)
		},
	}
}

func templatesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "templates",
		Short: "List stack templates",
		Run: func(cmd *cobra.Command, args []string) {
			var rows [][]string
			for _, template := range templates.Builtins() {
				rows = append(rows, []string{template.ID, template.Name, strings.Join(template.Services, ", "), template.Description})
			}
			terminal.Table(cmd.OutOrStdout(), []string{"ID", "Name", "Services", "Description"}, rows)
		},
	}
}

func searchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search supported services",
		Run: func(cmd *cobra.Command, args []string) {
			query := strings.Join(args, " ")
			var rows [][]string
			for _, service := range services.NewCatalog().Search(query) {
				rows = append(rows, []string{service.ID, service.Name, string(service.Category), service.Image, strings.Join(service.Versions, ", ")})
			}
			terminal.Table(cmd.OutOrStdout(), []string{"ID", "Name", "Category", "Image", "Versions"}, rows)
		},
	}
}

func boolText(ok bool) string {
	if ok {
		return "ok"
	}
	return "missing"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func dedupe(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
