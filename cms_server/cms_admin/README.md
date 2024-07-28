# CMS Package 

This document describes the CMS package, a Go library for building content management systems. It provides functionalities for model registration, data persistence, and defining routes for your CMS application.

## Installation
Not a git repo yet...

## Usage
Here's a breakdown of how to set up the CMS package in your application:

### Prerequisites
- A configured logger (using `zerolog/logger`)
- An established database connection (using `gorm.io/gorm`)
- A running server instance (using `gorilla/mux`)

### Steps

1. **Import the CMS package**:
```go
    import (
        cms "your_project_path/cms_server"
    )
```

2. **Set up CMS after creating logger and database connection**:
```go
    // ... (Create logger and database connection)
    cfg := cms.Config{
    Logger: log.Logger,
    DB:     db.DB,
    }
    cms.Setup(&cfg)
```

3. **Register your models**:
Use the `cms.Register` function to register any data structures you want to manage through the CMS.
```go
    type MyModel struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
        // ... other fields
    }

    func main() {
        // ... (previous steps)
        cms.Register(&MyModel{})
    }
```

4. **Define CMS routes**:
The `cms.Routes` function allows you to define routes for your CMS functionality within your existing server router.
```go
    func main() {
        // ... (previous steps)

        // Define your server router (e.g., using gorilla/mux)
        router := mux.NewRouter()

        // Append CMS routes to the server router
        cms.Routes(router)

        // ... (Start server)
    }
```
*Important Note*: The provided example for `cms.Routes` demonstrates defining admin routes. Make sure to adjust these routes to exclude the admin endpoints if you don't want them exposed in your application.

### Example basic setup
```go
    package main

    import (
        cms "cms_server"
    )

    func main() {
        log = GetLogger()
        log.Info().Msg("Starting Test to CMS")

        db, err := LoadDB()
        if err != nil {
            log.Fatal().Err(err).Msg("Error loading database")
        }

        server, err := GetServer()
        if err != nil {
            log.Fatal().Err(err).Msg("Error starting server")
        }

        // Setup cms
        cfg := cms.Config{
            Logger: log.Logger,
            DB:     db.DB,
        }
        cms.Setup(&cfg)
        // Register models into cms
        cms.Register(&Primitive{})

        // Append cms routes to server
        cms.Routes(server.Router())

        err = server.ListenAndServe()
        if err != nil {
            log.Fatal().Err(err).Msg("Error starting server")
        }
    }
```