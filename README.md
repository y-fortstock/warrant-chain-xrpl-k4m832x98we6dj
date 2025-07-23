# chain-xrpl

## Usage

### Docker-based Workflow

All development, build, and run operations are performed via Docker and Makefile. **You do not need to install Go or any dependencies locally**—just Docker and Make.

### Running the Application

Build and run the application in Docker:

```
make run
```

This will:
- Build the Docker image for chain-xrpl
- Run the container on port 8099

To stop the container:
```
docker stop chain-xrpl
```

### Configuration

The service uses [Viper](https://github.com/spf13/viper) for configuration management. You can configure the log level and format via environment variables, YAML file, or CLI flag.

#### Environment Variables

Set the log level:

```
LOG_LEVEL=debug make run
```

Set the log format:

```
LOG_FORMAT=json make run
```

#### YAML Config

Edit `config.yaml` in the project root (or specify with `--config`). Example:

```yaml
log:
  level: info   # or debug, warn, error
  format: logfmt # or json
server:
  listen: ":8099"
```

#### CLI Flag

You can specify a custom config file with:

```
docker run -v /path/to/config.yaml:/app/config.yaml -p 8099:8099 chain-xrpl --config /app/config.yaml
```

**Precedence:** CLI flag > environment variable > YAML file > default (info).

---

## Development

### Prerequisites
- [Docker](https://www.docker.com/)
- [Make](https://www.gnu.org/software/make/)

### Dependency Installation

Install Go dependencies and vendor them (in Docker):

```
make deps
```

### All Code Generation

To update submodules, install dependencies, and generate all code:

```
make regen
```

### Building and Running

Build the Docker image:

```
make build
```

Run the application:

```
make run
```

Full rebuild (regen + build):

```
make rebuild
```

---

## Project Structure

- `Dockerfile` — production build (multi-stage, Go + Alpine)
- `Dockerfile.make` — dev/make environment (buf, wire, protoc-gen-go, etc.)
- `.dockerignore` — excludes binaries, generated files, VCS, logs, node_modules, etc. from Docker context
- `Makefile` — all build, run, and codegen commands (see `make help`)
- `proto/` — protobuf definitions and generated code (standalone Go)
  - `go.mod` — proto is a Go
  - `blockchain/*/v1/` — proto and generated files for account, token, types
- `internal/di` — dependency injection (Google Wire)
- `config.yaml` — default config

---

## Protobuf as Submodule

You can add proto files as a git submodule if needed:

```
git submodule add <repo_with_proto> proto
```

---

## Notes
- **prototool** is no longer used for code generation; only **buf** is supported.
- All commands are run in Docker; local Go toolchain is not required.
- For available commands, run:

```
make help
```
