# Go App Builder v1.4.3

This is a Go library that provides a foundation for building applications. It offers functionalities for:

- [**Configuration Management**](#configuration-management): Loads configuration options from a YAML file.
- [**Logging**](#logger): Manages application logs for debugging and monitoring purposes.
- [**Database**](#database): Establishes connections to databases using GORM.
- [**Server**](#server): Provides a basic HTTP server with routing capabilities.
- [**Admin**](#admin):
- [**Firebase**]:

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
  ConfigFile: &builder.ConfigFile{
    UseConfigFile: true,  // Optional, defaults to false
    ConfigPath:    "config.yaml", // Optional, defaults to "config.yaml"
  },
}

builder := builder.NewBuilder(config)
```

---

## Configuration Management

The builder uses the `viper` library to manage application configuration loaded from a YAML file (default: `config.yaml`).
If you pass configuration settings on [builder initialization](#getting-started), you may have access to a `viper` instance.

```go
ConfigFile: &builder.ConfigFile{
  UseConfigFile: true,  // Optional, defaults to false
  ConfigPath:    "config.yaml", // Optional, defaults to "config.yaml"
}
```

### Accessing Configuration Values:

This method retrieves the underlying `viper` instance used by the builder. You can then use the various methods provided by `viper` to access configuration values:

- `Get(key string) interface{}`: Returns the value for the given key as an interface{}. You might need to type-cast it to the desired type.
- `GetString(key string) string`: Returns the value for the given key as a string.
- `GetInt(key string) int`: Returns the value for the given key as an integer.
- `GetFloat64(key string) float64`: Returns the value for the given key as a float64.
- `GetBool(key string) bool`: Returns the value for the given key as a bool.

```go
configReader := builder.GetConfigReader()
firebaseSecret := configReader.GetString("firebaseSecret")
```

### Reference

Refer to the viper documentation for a complete list of available methods: https://github.com/spf13/viper

---

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
loggerConfig := builder.LoggerConfig{
  LogLevel:    "debug",
  LogFilePath: "logs/default.log",
  WriteToFile: true,
}
engine.SetLoggerConfig(loggerConfig)

// Logging example
log, err = engine.GetLogger()
if err != nil {
  // handle error
}
log.Info().Msg("Some logging test")
```

### Reference

For more information on zerolog and its advanced features, refer to the official documentation: https://github.com/rs/zerolog

---

## Database

The builder library provides functionalities for connecting to databases using the GORM library.

### Database configuration

```go
type DBConfig struct {
    // URL: Used for connecting to a PostgreSQL database.
    // Provide a complete connection string (e.g., "postgres://user:password@host:port/database").
    URL string
    // Path: Used for connecting to a SQLite database.
    // Provide the path to the SQLite database file.
    Path string
}
```

Note: You can only specify one connection method (either `URL` or `Path`) at a time.

### Establishing a Database Connection

To establish a connection to the database, use the `builder.ConnectDB` method:

```go
dbConfig := builder.DBConfig{
    // URL:  "postgres://user:password@host:port/database",
    Path: cfg.GetString("dbFile"), // Example using a config file
}
err := builder.ConnectDB(&dbConfig)
if err != nil {
    // Handle error
}
```

### Reference

For more information refer to [GORM documentation](https://gorm.io/)

---

## Server

The builder library provides a basic HTTP server with routing capabilities, middleware support, and basic configuration options.

### Server configuration

You can configure the server host and port using the builder.ServerConfig struct:

```go
// Server setup
serverConfig := builder.ServerConfig{
  Host: cfg.GetString("host"),
  Port: cfg.GetString("port"),
}
err = engine.SetServerConfig(serverConfig)
if err != nil {
  // handle error
}
```

Note: If not configured, the server will default to listening on all interfaces (`0.0.0.0`) and port `8080`.

### Retrieve the server instance

You can access the server instance using the `builder.GetServer` method:

```go
svr, err := engine.GetServer()
if err != nil {
  // handle error
}
```

### Adding middlewares

Middleware allows you to intercept requests and responses, adding functionalities like logging, authentication, or request validation before reaching the actual route handler. You can chain multiple middleware functions.

```go
svr.AddMiddleware(func(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    log.Info().Msg(r.RequestURI)
    next.ServeHTTP(w, r)
  })
})
```

### Adding routes

```go
svr.AddRoute("/", func(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Home")
}, "home")
```

### Start the server

Start the server listening for requests with the `svr.Run` method:

```go
svr.Run()
```

### Reference

For extra documentation, visit [Mux](https://github.com/gorilla/mux)

---

## Admin

Once you have initialized your server, you can setup your admin panel

```go
	svr, err := engine.GetServer()
	if err != nil {
    // handle
	}

  ...

	// Admin setup --> Needs to happen after the server is setup
	err = engine.SetupAdmin()
	if err != nil {
		log.Error().Err(err).Msg("Error setting up admin panel")
		panic(err)
	}

	admin := engine.GetAdmin()
	admin.Register(&Example{})

  ...

	svr.Run()
```

This will setup db migration for that entity
and also the endpoints for crud operations

- list: `/`
- new: `/new`
- details: `/{id}`
- edit: `/{id}/edit`
- delete: `/{id}/delete`

---

## Firebase

Have configured `firebaseSecret` which should be a base64 encoding of the secret provided by google.
call the method `builder.SetupFirebase()`
