# AGENTS.md

This document provides essential context for agents working in the `stocks_broadcaster` repository.

## Project Overview

`stocks_broadcaster` is a Go application that subscribes to stock price data from Tinkoff's InvestAPI via gRPC and broadcasts it to Redis channels. It supports multiple input connections (to bypass API limits), multiple Redis outputs, and can push metrics to VictoriaMetrics. The application is designed to run as a systemd service and includes Kubernetes readiness/liveness probes.

## Essential Commands

### Development
- `make tools`: Verify required tools (podman, redis-cli, go) are installed and working
- `make deps`: Download and verify Go modules (`go mod download && go mod verify && go mod tidy`)
- `make test`: Run all Go tests (`go test -v ./...`)
- `make start`: Run the application with a config file (`go run main.go ./contrib/local.yaml`)
- `make vuln`: Run vulnerability scanning with `govulncheck` (if installed)

### Building
- `make build`: Build the production binary with version information
- `make binary`: Run the built binary with a config file

### Container Development
- `make docker/up`: Start development environment with Docker Compose (Redis, VictoriaMetrics, and app)
- `make docker/resource`: Start only the Redis and VictoriaMetrics containers
- `make docker/down`: Stop the development environment
- `make docker/prune`: Remove all development containers and volumes
- `make podman/up`: Same as docker/up but using podman
- `make podman/resource`: Start Redis and VictoriaMetrics with podman
- `make podman/down`: Stop podman environment
- `make podman/prune`: Remove podman containers and volumes

### Release
- `make tag`: Create a git tag with the current version (v<major>.<minor>.<patch>)

## Code Organization

- `main.go`: Application entry point - initializes config, logging, readers, writers, and starts the service
- `config/`: Configuration loading and structure definitions (YAML)
- `internal/model/`: Data models (Instrument, Update)
- `internal/service/`: Core business logic (Broadcaster service that coordinates readers and writers)
- `internal/transport/`: Protocol implementations
  - `reader/`: Input sources (currently only InvestAPI)
  - `writer/`: Output destinations (currently only Redis)
  - `http/`: HTTP server for health checks and metrics
  - `victoria_metrics/`: Metrics exporter to VictoriaMetrics
- `contrib/`: Example configuration files and systemd service definition

## Architecture

The application follows a publisher-subscriber pattern:

1. **Readers** (currently only InvestAPI) subscribe to stock price updates and send them to a central channel
2. **Broadcaster service** receives updates and routes them to configured **writers**
3. **Writers** (currently only Redis) publish the updates to their respective destinations

The `Broadcaster` struct in `internal/service/service.go` is the central coordination point. It maintains:
- A channel (`Cord`) for receiving updates from readers
- Maps for routing (FIGI to channel name, FIGI to display name)
- References to all readers and writers
- A subscription system for writers to receive updates

## Key Patterns and Conventions

### Configuration
- Configuration is loaded from YAML files
- The `Config` struct in `config/config.go` defines the schema with validation tags
- Example configuration in `contrib/stocks_broadcaster_example.yaml`
- Required configuration files must be provided as command-line arguments

### Error Handling
- Fatal errors use `log.Fatal()` with detailed context
- Recoverable errors are handled with standard Go error returns
- The application uses a context-based shutdown with a 10-second timeout

### Logging
- Uses `github.com/rs/zerolog` with configurable levels (trace, debug, info, warn, error)
- Can optionally log to journald
- Log levels control verbosity (trace for maximum detail)

### Testing
- Standard Go testing with `go test`
- Configuration loading has unit tests
- No integration tests visible in the codebase

## Gotchas and Non-Obvious Patterns

### Multiple Network Interfaces
The application can use different local network interfaces for different API connections by setting `local_addr` in the input configuration. This allows bypassing per-IP connection limits.

### Systemd Integration
The application integrates with systemd for service management:
- Uses `github.com/coreos/go-systemd/v22` for watchdog notifications
- If systemd is not available, it logs a warning but continues
- The watchdog is enabled automatically when running under systemd

### VictoriaMetrics Integration
- Metrics are pushed to VictoriaMetrics endpoints at configurable intervals
- Two time series are generated for each instrument:
  - `stocks_last_deal{...}`: The last deal price with lot size
  - `stocks_lot_price{...}`: The price per lot
- Can expose Go runtime metrics to VictoriaMetrics
- The `insecure_vm` config option disables TLS certificate verification for VictoriaMetrics endpoints

### HTTP Server
- Provides three endpoints:
  - `/ready`: Always returns 200 - indicates the process is running
  - `/healthz`: Checks service health by pinging all components
  - `/metrics`: Exposes Prometheus-formatted metrics (application and optionally Go runtime)
- Also exposes pprof endpoints for profiling at `/debug/pprof/`

### Concurrency Model
- Uses `golang.org/x/sync/errgroup` to manage multiple goroutines
- The main function starts several goroutines:
  - Readers
  - Writers
  - Systemd watchdog
  - HTTP server
  - VictoriaMetrics pushers
- All goroutines are cancelled when the main context is cancelled

### Build Versioning
- Version is injected at build time via `-ldflags "-X main.Version=$(ver)"`
- The version format is `<major>.<minor>.<patch>.<git_sha>`
- Development builds have version "development"

## Security Considerations
- API tokens are stored in configuration files - should be protected
- Redis connections use standard Redis URLs (can include authentication)
- VictoriaMetrics connections can use headers for authentication (e.g., Basic auth)
- TLS certificate verification for VictoriaMetrics can be disabled (insecure_vm) - use with caution

## Deployment
- Can be deployed via Docker using the provided Dockerfiles
- Production image is minimal (Alpine-based, ~40MB)
- Development image includes the Go toolchain
- Example systemd service file in `contrib/stocks_broadcaster.service`
- Can be orchestrated with Docker Compose for development

## Testing the Application
1. Configure `contrib/local.yaml` with valid API tokens and FIGIs
2. Start Redis (via `make docker/resource`, `make podman/resource`, or system service)
3. Run `make start`
4. Monitor Redis with `redis-cli monitor` to see published messages
5. Check `/healthz` and `/metrics` endpoints if webserver is enabled

The application will log subscription successes and price updates at appropriate log levels.