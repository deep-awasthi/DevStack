package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const FileName = "devstack.yml"

type Store struct {
	Path string
}

func NewStore(path string) Store {
	if path == "" {
		path = FileName
	}
	return Store{Path: path}
}

func (s Store) Exists() bool {
	_, err := os.Stat(s.Path)
	return err == nil
}

func (s Store) Load() (Environment, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return Environment{}, explainFileError("read", s.Path, err)
	}
	var env Environment
	if err := yaml.Unmarshal(data, &env); err != nil {
		return Environment{}, fmt.Errorf("could not parse %s: %w\nSolution: validate the YAML or regenerate it with devstack init", s.Path, err)
	}
	if env.Services == nil {
		env.Services = map[string]ServiceConfig{}
	}
	if env.Network == "" {
		env.Network = "devstack_" + sanitizeName(env.Name)
	}
	return env, nil
}

func (s Store) Save(env Environment) error {
	if env.Services == nil {
		env.Services = map[string]ServiceConfig{}
	}
	if env.Network == "" {
		env.Network = "devstack_" + sanitizeName(env.Name)
	}
	data, err := yaml.Marshal(env)
	if err != nil {
		return fmt.Errorf("could not serialize environment: %w", err)
	}
	if err := os.MkdirAll(absOrDot(s.Path), 0o755); err != nil {
		return explainFileError("create directory for", s.Path, err)
	}
	if err := os.WriteFile(s.Path, data, 0o600); err != nil {
		return explainFileError("write", s.Path, err)
	}
	return nil
}

func WriteJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func explainFileError(action, path string, err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not %s %s: file does not exist\nSolution: run devstack init or pass --config with an existing file", action, path)
	}
	if errors.Is(err, os.ErrPermission) {
		return fmt.Errorf("could not %s %s: permission denied\nSolution: check file ownership and directory permissions", action, path)
	}
	return fmt.Errorf("could not %s %s: %w", action, path, err)
}

func absOrDot(path string) string {
	dir := filepath.Dir(path)
	if dir == "" {
		return "."
	}
	return dir
}
