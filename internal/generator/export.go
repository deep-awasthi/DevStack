package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	appconfig "github.com/deepawasthi/devstack/internal/config"
	"github.com/deepawasthi/devstack/internal/services"
	"gopkg.in/yaml.v3"
)

type Exporter struct {
	Dir string
}

func NewExporter(dir string) Exporter {
	if dir == "" {
		dir = "."
	}
	return Exporter{Dir: dir}
}

func (e Exporter) Export(env appconfig.Environment, resolved []services.ResolvedService) error {
	if err := os.MkdirAll(e.Dir, 0o755); err != nil {
		return err
	}
	files := map[string][]byte{
		"docker-compose.yml":          e.compose(env, resolved),
		"docker-compose.override.yml": []byte("services: {}\n"),
		".env":                        []byte(e.dotenv(resolved)),
		"application.yml":             []byte(e.springYAML(resolved)),
		"application.properties":      []byte(e.springProperties(resolved)),
		"appsettings.json":            []byte(e.appsettings(resolved)),
		"README.md":                   []byte(e.readme(env, resolved)),
	}
	envData, err := yaml.Marshal(env)
	if err != nil {
		return err
	}
	files["devstack.yml"] = envData
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(e.Dir, name), data, 0o600); err != nil {
			return fmt.Errorf("could not write %s: %w", name, err)
		}
	}
	return nil
}

func (e Exporter) compose(env appconfig.Environment, resolved []services.ResolvedService) []byte {
	root := map[string]any{"name": env.Name, "services": map[string]any{}, "networks": map[string]any{env.Network: map[string]any{"name": env.Network}}, "volumes": map[string]any{}}
	serviceMap := root["services"].(map[string]any)
	volumeMap := root["volumes"].(map[string]any)
	for _, item := range resolved {
		entry := map[string]any{"image": item.Image, "container_name": item.Container, "restart": "unless-stopped", "networks": []string{env.Network}}
		if len(item.Service.Ports) > 0 {
			var bindings []string
			for _, port := range item.Service.Ports {
				bindings = append(bindings, fmt.Sprintf("%d:%d", item.Ports[port.Name], port.Internal))
			}
			entry["ports"] = bindings
		}
		if envs := services.Environment(item); len(envs) > 0 {
			entry["environment"] = envs
		}
		if cmd := services.Command(item); len(cmd) > 0 {
			entry["command"] = cmd
		}
		if len(item.Service.Volumes) > 0 {
			var volumes []string
			for _, volume := range item.Service.Volumes {
				name := item.VolumeNames[volume.Name]
				volumes = append(volumes, name+":"+volume.Path)
				volumeMap[name] = map[string]any{"name": name}
			}
			entry["volumes"] = volumes
		}
		serviceMap[item.Service.ID] = entry
	}
	data, _ := yaml.Marshal(root)
	return data
}

func (e Exporter) dotenv(resolved []services.ResolvedService) string {
	var lines []string
	for _, item := range resolved {
		prefix := strings.ToUpper(strings.ReplaceAll(item.Service.ID, "-", "_"))
		for _, port := range item.Service.Ports {
			lines = append(lines, fmt.Sprintf("%s_%s_PORT=%d", prefix, strings.ToUpper(port.Name), item.Ports[port.Name]))
		}
		for key, value := range item.Credentials {
			lines = append(lines, fmt.Sprintf("%s_%s=%s", prefix, strings.ToUpper(key), value))
		}
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n") + "\n"
}

func (e Exporter) springYAML(resolved []services.ResolvedService) string {
	values := map[string]any{"spring": map[string]any{}}
	spring := values["spring"].(map[string]any)
	for _, item := range resolved {
		switch item.Service.ID {
		case "postgres":
			spring["datasource"] = map[string]any{"url": fmt.Sprintf("jdbc:postgresql://localhost:%d/%s", item.Ports["postgres"], item.Credentials["database"]), "username": item.Credentials["username"], "password": item.Credentials["password"]}
		case "redis":
			spring["data"] = map[string]any{"redis": map[string]any{"host": "localhost", "port": item.Ports["redis"], "password": item.Credentials["password"]}}
		case "mongo":
			spring["data"] = map[string]any{"mongodb": map[string]any{"uri": services.Render("mongodb://{{username}}:{{password}}@localhost:{{port:mongo}}/{{database}}?authSource=admin", item)}}
		}
	}
	data, _ := yaml.Marshal(values)
	return string(data)
}

func (e Exporter) springProperties(resolved []services.ResolvedService) string {
	var lines []string
	for _, item := range resolved {
		switch item.Service.ID {
		case "postgres":
			lines = append(lines, fmt.Sprintf("spring.datasource.url=jdbc:postgresql://localhost:%d/%s", item.Ports["postgres"], item.Credentials["database"]))
			lines = append(lines, "spring.datasource.username="+item.Credentials["username"])
			lines = append(lines, "spring.datasource.password="+item.Credentials["password"])
		case "redis":
			lines = append(lines, "spring.data.redis.host=localhost", fmt.Sprintf("spring.data.redis.port=%d", item.Ports["redis"]), "spring.data.redis.password="+item.Credentials["password"])
		case "mongo":
			lines = append(lines, "spring.data.mongodb.uri="+services.Render("mongodb://{{username}}:{{password}}@localhost:{{port:mongo}}/{{database}}?authSource=admin", item))
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func (e Exporter) appsettings(resolved []services.ResolvedService) string {
	var lines []string
	lines = append(lines, "{", `  "ConnectionStrings": {`)
	for _, item := range resolved {
		if item.Service.ID == "postgres" {
			lines = append(lines, fmt.Sprintf(`    "Postgres": "Host=localhost;Port=%d;Database=%s;Username=%s;Password=%s"`, item.Ports["postgres"], item.Credentials["database"], item.Credentials["username"], item.Credentials["password"]))
		}
	}
	lines = append(lines, "  }", "}", "")
	return strings.Join(lines, "\n")
}

func (e Exporter) readme(env appconfig.Environment, resolved []services.ResolvedService) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s DevStack\n\n", env.Name)
	b.WriteString("Generated by DevStack.\n\n## Services\n\n")
	for _, item := range resolved {
		fmt.Fprintf(&b, "- %s `%s`\n", item.Service.Name, item.Image)
		for _, port := range item.Service.Ports {
			fmt.Fprintf(&b, "  - %s: localhost:%d\n", port.Name, item.Ports[port.Name])
		}
	}
	b.WriteString("\n## Commands\n\n")
	b.WriteString("- `devstack up`\n- `devstack down`\n- `devstack logs <service>`\n- `devstack inspect`\n")
	return b.String()
}
