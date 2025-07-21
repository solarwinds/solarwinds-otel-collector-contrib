# Registry

This module provides a simple interface for accessing Windows Registry keys and values on Windows systems.

## Features

- Convenience direct reading functions
- Support for string and integer value reading
- Registry key enumeration

## Usage

### Basic Registry Reading

```go
func main() {
    // Create a registry reader for HKEY_LOCAL_MACHINE
    reader, err := registry.NewReader(
        registry.LocalMachineKey,
        "SYSTEM\\CurrentControlSet\\Services\\SharedAccess\\Parameters\\FirewallPolicy\\StandardProfile",
    )
    if err != nil {
        log.Fatal(err)
    }

    // Read a DWORD value
    enabled, err := reader.GetKeyUIntValue("", "EnableFirewall")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Firewall enabled: %t\n", enabled == 1)

    // Read string values
    values, err := reader.GetKeyValues("", []string{"ProgramFilesDir", "CommonFilesDir"})
    if err != nil {
        log.Fatal(err)
    }
    for name, value := range values {
        fmt.Printf("%s: %s\n", name, value)
    }
}
```

### Using Convenience Functions

```go
func main() {
    // Direct access without creating a reader instance
    values, err := registry.GetKeyValues(
        registry.CurrentUserKey,
        "Control Panel",
        "International",
        []string{"LocaleName", "Locale"},
    )
    if err != nil {
        log.Fatal(err)
    }

    for name, value := range values {
        fmt.Printf("%s: %s\n", name, value)
    }
}
```

## Platform Support

This package is designed for Windows systems only. Returns error on other platforms. The Windows-specific implementation is in [`registryreader_windows.go`](registryreader_windows.go).

## Dependencies

- `golang.org/x/sys/windows/registry`: Core Windows Registry functionality
