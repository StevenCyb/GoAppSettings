# 🔧 GoAppSettings

**A powerful, type-safe configuration loader for Go applications with layered configuration sources.**

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue.svg)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-95.1%25-green.svg)](https://github.com/StevenCyb/GoAppSettings)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## ✨ Features

- 🎯 **Type-safe** configuration using Go generics
- 📁 **Multiple config sources** with clear priority hierarchy
- 🔄 **Automatic type conversion** (string, int, float, bool)
- 🏗️ **Builder pattern** for easy configuration
- 🧪 **Fully tested** with 95%+ code coverage
- 🚀 **Zero dependencies** (only standard library)

## 🏆 Configuration Priority Hierarchy

```
┌─────────────────────────────────────┐
│            Command Line Args        │  ← Highest Priority
│              --port 3000            │
├─────────────────────────────────────┤
│         Environment Variables       │
│           PORT=8080                 │
├─────────────────────────────────────┤
│      Environment Config File        │
│        config.dev.json              │
├─────────────────────────────────────┤
│         Base Config File            │  ← Lowest Priority
│          config.json                │
└─────────────────────────────────────┘
```

## 📦 Installation

```bash
go get github.com/StevenCyb/GoAppSettings
```

## 🚀 Quick Start

### 1. Define Your Configuration Structure

```go
package main

import (
    "fmt"
    "os"
    "github.com/StevenCyb/GoAppSettings/pkg/appsettings"
)

type Config struct {
    DatabaseURL string  `json:"databaseURL"`
    Port        int     `json:"port"`
    DebugMode   bool    `json:"debugMode"`
    Timeout     float64 `json:"timeout"`
}
```

### 2. Load Configuration

```go
func main() {
    config, err := appsettings.New[Config]().
        //  Use command line arguments
        WithArgs(os.Args).
        //  Specify config directory (default is current working directory)
        WithConfigDirectory("./config").
        //  Use environment variables
        WithEnvVars(os.Environ()).
        // Specify the environment (e.g., "dev", "prod", "test") (default not used)
        WithEnvironment("dev").
        // Finally, load the configuration
        Load()
}
```

## 📋 Complete Usage Examples

### Example 1: Base Configuration Only

**config.json:**
```json
{
    "databaseURL": "postgres://localhost/my_app",
    "port": 8080,
    "debugMode": false,
    "timeout": 30.0
}
```

**Command:**
```bash
go run main.go
```

**Result:**
```go
&Config{
    DatabaseURL: "postgres://localhost/my_app",
    Port:        8080,
    DebugMode:   false,
    Timeout:     30.0,
}
```

---

### Example 2: Environment-Specific Override

**config.json:**
```json
{
    "databaseURL": "postgres://localhost/my_app",
    "port": 8080,
    "debugMode": false,
    "timeout": 30.0
}
```

**config.dev.json:**
```json
{
    "databaseURL": "postgres://dev-server/my_app_dev",
    "debugMode": true,
}
```

**Command:**
```bash
go run main.go
```

**Result:**
```go
&Config{
    DatabaseURL: "postgres://dev-server/my_app_dev", // ← Overridden by dev config
    Port:        8080,                              // ← From base config
    DebugMode:   true,                              // ← Overridden by dev config
    Timeout:     30.0,                              // ← Kept from base config
}
```

---

### Example 3: Environment Variables Override

**With config files from Example 2...**

**Command:**
```bash
PORT=9000 TIMEOUT=60.5 go run main.go
```

**Result:**
```go
&Config{
    DatabaseURL: "postgres://dev-server/my_app_dev", // ← From config file
    Port:        9000,                              // ← Overridden by env var
    DebugMode:   true,                              // ← From config file
    Timeout:     60.5,                              // ← Overridden by env var
}
```

---

### Example 4: Command Line Arguments (Highest Priority)

**With config files and env vars from previous examples...**

**Command:**
```bash
PORT=9000 TIMEOUT=60.5 go run main.go --port 3000 --debugmode false
```

**Result:**
```go
&Config{
    DatabaseURL: "postgres://dev-server/my_app_dev", // ← From dev config
    Port:        3000,                              // ← Overridden by CLI arg (highest priority!)
    DebugMode:   false,                             // ← Overridden by CLI arg
    Timeout:     60.5,                              // ← From env var
}
```

---

### Example 5: Complete Override Demonstration

Let's trace how a single field gets its final value:

**Configuration Sources:**

```bash
# config.json
{"port": 8080}

# config.dev.json  
{"port": 8090}

# Environment Variable
PORT=9000

# Command Line Argument
--port 3000
```

**Override Chain:**
```
Port Value Evolution:
8080 (base config) 
  ↓
8090 (dev config overrides base)
  ↓  
9000 (env var overrides dev config)
  ↓
3000 (CLI arg overrides env var) ← Final Value
```

## 🔧 Advanced Usage

### Custom Configuration Directory

```go
config, err := appsettings.New[Config]().
    WithConfigDirectory("/etc/my_app").  // Custom config directory
    WithEnvironment("production").
    Load()
```

### Environment Variable Mapping

Environment variables are automatically mapped to JSON field names using case-insensitive matching:

```bash
# These environment variables (CORRECT format)...
DATABASEURL=postgres://prod/db    # Matches json:"databaseURL" (case-insensitive)
# OR
databaseurl=postgres://prod/db    # Also matches json:"databaseURL" (case-insensitive)
DEBUGMODE=true                   # Matches json:"debugMode" (case-insensitive)
# OR  
debugmode=true                   # Also matches json:"debugMode" (case-insensitive)
PORT=5432                        # Matches json:"port" (exact match)

# Map to these JSON fields...
{
    "databaseURL": "postgres://prod/db",
    "debugMode": true,
    "port": 5432
}

# These environment variables (INCORRECT format) will be IGNORED...
DATABASE_URL=postgres://prod/db  # ❌ Doesn't match json:"databaseURL" 
DEBUG_MODE=true                  # ❌ Doesn't match json:"debugMode"
```

### Command Line Argument Formats

```bash
# Key-value pairs
--port 8080
--databaseurl postgres://localhost/db  # Matches json:"databaseURL"

# Boolean flags (automatically set to true)
--debugmode                            # Matches json:"debugMode"
--verbose

# Mixed usage
go run main.go --port 3000 --debugmode --timeout 45.5
```

### ✅ Correct vs ❌ Incorrect Usage Examples

```bash
# ✅ CORRECT - Environment variables match JSON field names (any case)
DATABASEURL=postgres://prod/db DEBUGMODE=true PORT=8080 go run main.go
# OR (lowercase also works)
databaseurl=postgres://prod/db debugmode=true port=8080 go run main.go

# ❌ INCORRECT - Underscores don't match camelCase JSON tags  
DATABASE_URL=postgres://prod/db DEBUG_MODE=true PORT=8080 go run main.go
# Result: DATABASE_URL and DEBUG_MODE are ignored, fields remain at default values

# ✅ CORRECT - Command line args
go run main.go --databaseurl postgres://prod/db --debugmode true --port 8080

# ❌ INCORRECT - Using underscores or hyphens inconsistently
go run main.go --database-url postgres://prod/db --debug-mode true --port 8080  
# Result: Arguments are converted to "database-url" and "debug-mode" which don't match JSON tags
```

## 📁 File Structure

```
your-app/
├── config.json           # Base configuration
├── config.dev.json       # Development overrides
├── config.prod.json      # Production overrides
├── config.test.json      # Testing overrides
└── main.go
```

## 🎯 Type Conversion

GoAppSettings automatically converts string values to appropriate types:

| Input String | Detected Type | Go Value |
|--------------|---------------|----------|
| `"true"` | `bool` | `true` |
| `"false"` | `bool` | `false` |  
| `"123"` | `int` | `123` |
| `"45.67"` | `float64` | `45.67` |
| `"hello"` | `string` | `"hello"` |

## 🏗️ Builder Methods

| Method | Description | Example |
|--------|-------------|---------|
| `WithArgs([]string)` | Set command line arguments | `.WithArgs(os.Args)` |
| `WithEnvVars([]string)` | Set environment variables | `.WithEnvVars(os.Environ())` |
| `WithEnvironment(string)` | Set environment name for config files | `.WithEnvironment("dev")` |
| `WithConfigDirectory(string)` | Set custom config directory | `.WithConfigDirectory("/etc/app")` |

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE.txt) file for details.
