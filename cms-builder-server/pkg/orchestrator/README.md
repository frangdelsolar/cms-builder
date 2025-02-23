# Orchestrator

Will initialize

## Config Reader

The builder uses the `viper` library to manage application configuration loaded from a YAML file (default: `config.yaml`).
If you pass configuration settings on [builder initialization](#getting-started), you may have access to a `viper` instance.

### Accessing Configuration Values:

This method retrieves the underlying `viper` instance used by the builder. You can then use the various methods provided by `viper` to access configuration values:

- `Get(key string) interface{}`: Returns the value for the given key as an interface{}. You might need to type-cast it to the desired type.
- `GetString(key string) string`: Returns the value for the given key as a string.
- `GetInt(key string) int`: Returns the value for the given key as an integer.
- `GetFloat64(key string) float64`: Returns the value for the given key as a float64.
- `GetBool(key string) bool`: Returns the value for the given key as a bool.

`environment.go` has a `EnvKeys` var that should contain the same keys in the environment.
this assures safetu when getting the variables

```go
	o, err := orc.NewOrchestrator()
	config := o.Config
	appName := config.GetString(orc.EnvKeys.AppName)
```

## Logger
