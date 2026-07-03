# DevStack

DevStack is a terminal-based Docker environment manager for backend developers. It helps create, run, inspect, and export local development infrastructure without manually writing Docker commands or Docker Compose files.

The project is written in Go and uses Cobra for the CLI, Bubble Tea for interactive terminal UI, Lip Gloss for styling, Viper for configuration, and Docker CLI through an internal runtime abstraction.

## Features

- Interactive searchable service selector
- Docker installation, engine, Compose, and Podman detection
- Built-in service catalog for databases, brokers, observability tools, local cloud services, gateways, caches, and developer utilities
- Ready-made stack templates
- `devstack.yml` environment persistence
- Secure credential generation
- Automatic local port allocation
- Docker image detection and pull-on-demand
- Docker network, volume, and container management
- Health, stats, logs, and connection details
- Docker Compose and application configuration export
- Runtime abstraction designed for a future Docker SDK or Podman backend

## Requirements

- Go 1.25 or newer
- Docker CLI
- Docker Engine or Docker Desktop
- Docker Compose plugin

Podman detection is included for future runtime support.

## Install

```sh
go build -o devstack .
```

Optionally move the binary somewhere on your `PATH`:

```sh
mv devstack /usr/local/bin/devstack
```

## Quick Start

Create an environment from a template:

```sh
devstack init --name ecommerce --template spring-boot-starter
```

Start the environment:

```sh
devstack up
```

Inspect services and connection details:

```sh
devstack inspect
```

Export Docker Compose and app configuration:

```sh
devstack export --dir ./devstack-export
```

Stop services:

```sh
devstack down
```

## Interactive Mode

Run DevStack with no subcommand:

```sh
devstack
```

The interactive terminal UI checks Docker status, then opens a searchable multi-select service picker.

Controls:

- `SPACE` selects or deselects a service
- `/` searches services
- `ENTER` continues
- `q` quits

## Commands

```text
devstack
devstack init
devstack up
devstack down
devstack stop
devstack restart
devstack remove
devstack logs
devstack connect
devstack inspect
devstack stats
devstack list
devstack cleanup
devstack update
devstack export
devstack doctor
devstack version
devstack templates
devstack search
devstack help
```

## Templates

List available templates:

```sh
devstack templates
```

Included templates:

- `spring-boot-starter`
- `microservices`
- `node-api`
- `search-platform`
- `authentication`
- `cloud-development`
- `ai-development`
- `developer-essentials`
- `observability`

Use a template:

```sh
devstack init --template developer-essentials
```

## Search Services

```sh
devstack search redis
devstack search kafka
devstack search observability
```

## Configuration

DevStack persists environments in `devstack.yml`.

Example:

```yaml
name: ecommerce
network: devstack_ecommerce
createdWith: devstack
services:
  postgres:
    version: "17"
    enabled: true
    ports:
      postgres: 5432
    credentials:
      username: postgres
      database: app
      password: generated_password
  redis:
    version: "8"
    enabled: true
    ports:
      redis: 6379
    credentials:
      password: generated_password
```

Use a custom config path:

```sh
devstack --config ./infra/devstack.yml up
```

## Exported Files

`devstack export` generates:

- `docker-compose.yml`
- `docker-compose.override.yml`
- `.env`
- `application.yml`
- `application.properties`
- `appsettings.json`
- `README.md`
- `devstack.yml`

Generated files use the actual ports and credentials from the resolved environment.

## Development

Run tests:

```sh
go test ./...
```

Build:

```sh
go build ./...
```

Run doctor:

```sh
go run . doctor
```

## Architecture

The codebase is organized around clean boundaries:

- `cmd/` Cobra command wiring
- `internal/app/` orchestration and use cases
- `internal/engine/` Docker runtime interface and Docker CLI implementation
- `internal/services/` service catalog, metadata, resolution, credentials, rendering
- `internal/config/` environment persistence
- `internal/generator/` Compose and app config export
- `internal/templates/` reusable stack templates
- `internal/ui/` Bubble Tea terminal UI
- `internal/ports/` local port allocation
- `internal/terminal/` terminal table rendering
- `internal/utils/` shared utility helpers

## Current Status

DevStack is a production-oriented first version. It includes the full CLI surface, real Docker CLI integration, environment persistence, export generation, and an extensible service catalog. Some advanced service-specific orchestration, deep health checks, and Docker SDK/Podman backends are planned future enhancements.
