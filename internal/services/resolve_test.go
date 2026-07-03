package services

import (
	"testing"

	appconfig "github.com/deepawasthi/devstack/internal/config"
)

func TestResolveEnvironmentGeneratesPortsAndCredentials(t *testing.T) {
	env := appconfig.NewEnvironment("shop")
	env.Services["postgres"] = appconfig.ServiceConfig{Enabled: true, Version: "17"}
	env.Services["redis"] = appconfig.ServiceConfig{Enabled: true, Version: "8"}

	resolved, updated, err := ResolveEnvironment(env, NewCatalog())
	if err != nil {
		t.Fatalf("ResolveEnvironment() error = %v", err)
	}
	if len(resolved) != 2 {
		t.Fatalf("expected 2 resolved services, got %d", len(resolved))
	}
	if updated.Services["postgres"].Ports["postgres"] == 0 {
		t.Fatal("expected postgres port to be allocated")
	}
	if updated.Services["postgres"].Credentials["password"] == "" {
		t.Fatal("expected postgres password to be generated")
	}
	if updated.Services["redis"].Credentials["password"] == "" {
		t.Fatal("expected redis password to be generated")
	}
}

func TestCatalogSearchFindsByCategory(t *testing.T) {
	matches := NewCatalog().Search("message broker")
	if len(matches) == 0 {
		t.Fatal("expected broker services")
	}
	found := false
	for _, match := range matches {
		if match.ID == "rabbitmq" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected rabbitmq in broker search results")
	}
}
