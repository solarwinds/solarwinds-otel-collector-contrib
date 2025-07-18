# WMI Package

This package provides a simple interface for executing WMI (Windows Management Instrumentation) queries on Windows systems.

## Features

- Generic interface for WMI query execution
- Support for single result and multiple result queries
- Error handling for common WMI query scenarios
- Mock implementation for testing

## Usage

### Basic Query Execution

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/solarwinds/solarwinds-otel-collector-contrib/tools/wmi"
)

type SystemInfo struct {
    TotalPhysicalMemory uint64
    ComputerName        string
}

func main() {
    executor := wmi.NewExecutor()
    
    // Query for single result
    result, err := wmi.QuerySingleResult[SystemInfo](executor)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Computer: %s, Memory: %d\n", result.ComputerName, result.TotalPhysicalMemory)
}
```

### Query Multiple Results

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/solarwinds/solarwinds-otel-collector-contrib/tools/wmi"
)

type Process struct {
    Name string
    ProcessId uint32
}

func main() {
    executor := wmi.NewExecutor()
    
    // Query for multiple results
    processes, err := wmi.QueryResult[[]Process](executor)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, process := range processes {
        fmt.Printf("Process: %s (PID: %d)\n", process.Name, process.ProcessId)
    }
}
```

### Testing with Mock

```go
package main

import (
    "testing"
    
    "github.com/solarwinds/solarwinds-otel-collector-contrib/tools/wmi"
)

func TestWithMock(t *testing.T) {
    mockData := []interface{}{
        &[]Process{{Name: "test.exe", ProcessId: 1234}},
    }
    
    executor := wmi.CreateWmiExecutorMock(mockData, nil)
    
    result, err := wmi.QueryResult[[]Process](executor)
    if err != nil {
        t.Fatal(err)
    }
    
    // Test your logic here
}
```

## Error Types

The package defines several error types for common WMI query scenarios:

- `ErrWmiNoResult`: No result returned from query
- `ErrWmiEmptyResult`: Empty result returned from query  
- `ErrWmiTooManyResults`: More than one result returned when expecting single result

## Platform Support

This package is designed for Windows systems only. The Windows-specific implementation is in `wmi_windows.go`.

## Dependencies

- `github.com/yusufpapurcu/wmi`: Core WMI functionality
- `go.uber.org/zap`: Logging

## License

Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.

Licensed under the Apache License, Version 2.0.
