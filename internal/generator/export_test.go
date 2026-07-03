package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/deepawasthi/devstack/internal/config"
	"github.com/deepawasthi/devstack/internal/services"
)

func TestExportWritesComposeAndEnv(t *testing.T) {
	env := appconfig.NewEnvironment("shop")
	env.Services["postgres"] = appconfig.ServiceConfig{Enabled: true, Version: "17"}
	resolved, updated, err := services.ResolveEnvironment(env, services.NewCatalog())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	dir := t.TempDir()
	if err := NewExporter(dir).Export(updated, resolved); err != nil {
		t.Fatalf("export: %v", err)
	}
	compose, err := os.ReadFile(filepath.Join(dir, "docker-compose.yml"))
	if err != nil {
		t.Fatalf("read compose: %v", err)
	}
	if !strings.Contains(string(compose), "postgres:17") {
		t.Fatalf("compose missing postgres image:\n%s", compose)
	}
	dotenv, err := os.ReadFile(filepath.Join(dir, ".env"))
	if err != nil {
		t.Fatalf("read dotenv: %v", err)
	}
	if !strings.Contains(string(dotenv), "POSTGRES_PASSWORD=") {
		t.Fatalf(".env missing password:\n%s", dotenv)
	}
}
