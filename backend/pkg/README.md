# ServiceContext

## Overview

ServiceContext is a lightweight application container used to manage infrastructure components and control their lifecycle in the application.

It is designed to centralize initialization, configuration, and shutdown of shared components such as logging, configuration, HTTP server, worker pools, and other infrastructure services.

It does not manage business logic or application use cases directly.

---

## Core Concept

The ServiceContext acts as a registry and lifecycle manager for components.

Each component is responsible for:

* Declaring its identity
* Initializing its configuration flags
* Activating itself using the ServiceContext
* Handling graceful shutdown

---

## Component Interface

Each component in the system must implement the following interface:

```go
type Component interface {
    ID() string
    InitFlags()
    Activate(ServiceContext) error
    Stop() error
}
```

### Responsibilities

* ID: returns a unique identifier for the component
* InitFlags: registers command-line flags or configuration values
* Activate: initializes the component using dependencies from ServiceContext
* Stop: handles cleanup during application shutdown

---

## ServiceContext Interface

The ServiceContext provides access to registered components and global services.

```go
type ServiceContext interface {
    Load() error
    Stop() error

    Logger(prefix string) logger.Logger
    LogLevel() string

    EnvName() string
    GetName() string

    Get(id string) (interface{}, bool)
    MustGet(id string) interface{}

    OutEnv()
}
```

### Responsibilities

* Load: initializes all registered components
* Stop: gracefully shuts down all components
* Get / MustGet: retrieves components by ID
* Logger: provides structured logging
* Environment management: handles application environment configuration
* OutEnv: prints resolved environment configuration

---

## Lifecycle Flow

The ServiceContext manages the application lifecycle in the following order:

1. Create ServiceContext with registered components
2. Initialize flags for all components
3. Parse environment variables and configuration
4. Load all components (Activate)
5. Application runs
6. Stop all components gracefully on shutdown

---

## Component Registration

Components are registered during ServiceContext creation:

```go
serviceCtx := sctx.NewServiceContext(
    sctx.WithName("my-service"),
    sctx.WithComponent(configComponent),
    sctx.WithComponent(httpServerComponent),
    sctx.WithComponent(workerComponent),
)
```

Each component is stored internally and managed by the ServiceContext.

---

## Design Principles

* Infrastructure management is centralized
* Business logic is not managed by ServiceContext
* Components are independent and self-initializing
* Dependency access is provided through ServiceContext only when necessary
* Lifecycle (init and shutdown) is handled in a consistent order

---

## Notes

* ServiceContext is not a dependency injection framework
* It does not automatically resolve dependency graphs
* It acts as a controlled registry for infrastructure components
* Application composition is handled separately (e.g., via composer layer)
