# Orchestrator v1.6.48

Will initialize

# Use it through the command line

Install by running

```bash
go install github.com/frangdelsolar/cms-builder/cms-builder-server
```

This command will compile the application and place the binary in $GOPATH/bin.

Then run

```bash
cms-builder-server -env=prod -postman=true
```

## Config Reader

The builder uses the `viper` library to manage application configuration loaded from a YAML file (default: `config.yaml`).
If you pass configuration settings on [builder initialization](#getting-started), you may have access to a `viper` instance.

## Example of initialization with retry

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	orc "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/orchestrator"
	"github.com/joho/godotenv"
)

func main() {
	prepareEnvironment()

	e, err := orc.NewOrchestrator()
	if err != nil {
		panic(err)
	}

	startServer(e)

}

var (
	retryTimestamps []time.Time
	retryMutex      sync.Mutex
	retryLimit      = 5
	retryWindow     = 2 * time.Minute
)

func shouldRetry() bool {
	retryMutex.Lock()
	defer retryMutex.Unlock()

	// Remove timestamps older than the retry window
	now := time.Now()
	var validTimestamps []time.Time
	for _, t := range retryTimestamps {
		if now.Sub(t) <= retryWindow {
			validTimestamps = append(validTimestamps, t)
		} else {
			fmt.Printf("Removing old timestamp: %v\n", t)
		}
	}
	retryTimestamps = validTimestamps

	fmt.Printf("Retry timestamps: %v\n", retryTimestamps)

	// Check if the number of retries within the window exceeds the limit
	if len(retryTimestamps) >= retryLimit {
		fmt.Println("Retry limit exceeded")
		return false
	}

	// Add the current retry timestamp
	retryTimestamps = append(retryTimestamps, now)
	fmt.Println("Retrying...")
	return true
}

func startServer(e *orc.Orchestrator) {
	defer func() {
		if r := recover(); r != nil {
			// Log the panic and restart the server
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("panic: %v", r)
			}

			fmt.Println("Panic recovered, error:", err)
			e.Logger.Error().Err(err).Msg("Server panicked, checking retry logic...")

			// Check retry logic
			if shouldRetry() {
				time.Sleep(5 * time.Second)
				fmt.Println("Restarting server...")
				startServer(e)
			} else {
				fmt.Println("Retry limit exceeded, server not restarted")
				e.Logger.Error().Err(errors.New("retry limit exceeded")).Msg("Server cannot be restarted, retry limit exceeded")
			}
		}
	}()

	if err := e.Run(); err != nil {
		fmt.Println("Server run error:", err)
		e.Logger.Error().Err(err).Msg("Error running server")
	}
}
```

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
