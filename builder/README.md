# Go App Builder v0.0.1
This is a Go library that provides a foundation for building applications. It offers functionalities for:

- **Logging**: Manages application logs for debugging and monitoring purposes.
- **Configuration Management**: Loads configuration options from a YAML file.

### Getting Started
1. **Install dependencies**:
```bash
go get github.com/frangdelsolar/go-builder
```
2. **Import the library**:
```go
import "github.com/frangdelsolar/go-builder"
```
3. **Create a builder instance**:
```go
config := &builder.BuilderConfig{
  Environment: "dev", // Optional, defaults to "dev"
  LoggerConfig: &builder.LoggerConfig{
    LogLevel:    "debug", // Optional, defaults to "info"
    WriteToFile: true,   // Optional, defaults to true
    LogFilePath: "logs/app.log", // Optional, defaults to "logs/default.log"
  },
}

builder := builder.NewBuilder(config)
```

## Logger
The builder provides a pre-configured `zerolog` logger instance. It allows for centralized logging with customizable levels and output destination.
**Levels**: You can configure the logging level using the `LogLevel` field in the `LoggerConfig` struct. Supported levels are:
- `debug`
- `info`
- `warn`
- `error`
- `fatal`
**Output**: By default, logs are written to both console and a file (`logs/default.log`). You can disable writing to a file by setting `WriteToFile` to false in the `LoggerConfig`. You can also customize the log file path with `LogFilePath`.

### Accessing the Logger:
```go
logger := builder.GetLogger()
logger.Info().Msg("Application started")
```

## Configuration
The builder uses the `viper` library to manage application configuration loaded from a YAML file (default: `config.yaml`). You need to call `builder.LoadConfig()` to read the configuration file before accessing its values.

### Loading Configuration:
```go
err := builder.LoadConfig()
if err != nil {
  // Handle error
}
```

### Accessing Configuration Values:
You can either access the viper instance with the `builder.GetConfigReader()` method provided, or you can access a specific variable `builder.GetConfig(key: string)`.
```go
configReader := builder.GetConfigReader()
firebaseSecret := configReader.GetString("firebaseSecret")
```
