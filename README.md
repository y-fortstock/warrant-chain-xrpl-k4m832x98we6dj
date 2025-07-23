# chain-xrpl

## Usage

### Running the Application

You can run the gRPC server using the CLI:

```
go run ./cmd/chain-xrpl
```

Or build the binary and run it:

```
go build -o bin/chain-xrpl ./cmd/chain-xrpl
./bin/chain-xrpl
```

### Configuration

The service uses [Viper](https://github.com/spf13/viper) for configuration management. You can configure the log level and format via environment variables, YAML file, or CLI flag.

#### Environment Variables

Set the log level:

```
LOG_LEVEL=debug ./chain-xrpl
```

Set the log format:

```
LOG_FORMAT=json ./chain-xrpl
```

#### YAML Config

Create a `config.yaml` file in the project root (or specify with `--config`). Example:

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
./chain-xrpl --config /path/to/config.yaml
```

**Precedence:** CLI flag > environment variable > YAML file > default (info).

---

## Development

### Dependency Installation

Install Go dependencies and vendor them:

```
go mod tidy
go mod vendor
```

Or use the Makefile:

```
make deps
```

### Code Generation

#### Dependency Injection (Wire)

Generate DI code with Google Wire:

```
cd internal/di
wire
```

Or via Makefile:

```
make wire
```

#### Protobuf (buf)

Generate Go code from protobuf definitions:

```
cd proto
buf generate
```

Or via Makefile:

```
make proto
```

### Building and Running

Build the binary:

```
make build
```

Run the application:

```
make run
```

### All-in-One

To generate proto, install dependencies, generate wire, and build:

```
make all
```

### Protobuf as Submodule

It is recommended to add proto files as a git submodule:

```
git submodule add <repo_with_proto> proto
```
