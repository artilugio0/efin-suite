# Agent Guidelines for efin-suite

## Build Commands
- Build: `./scripts/buid.sh` (note: script has typo in filename)
- Clean build: `CGO_ENABLED=0 go build -o build/efin cmd/efin`

## Test Commands
- No test files currently exist in the codebase
- Run all tests: `go test ./...` (when tests are added)

## Code Style Guidelines

### Formatting
- Use `gofmt` for automatic formatting
- Follow standard Go formatting conventions

### Imports
- Group imports: standard library first, then third-party, then local packages
- Use blank lines between import groups
- Example:
```go
import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    tea "github.com/charmbracelet/bubbletea"

    "github.com/artilugio0/efin-suite/internal/ql"
)
```

### Naming Conventions
- Functions/variables: camelCase (e.g., `dbFile`, `doRequestQuery`)
- Exported types/functions: PascalCase (e.g., `QueryResultsView`, `Execute`)
- Struct fields: PascalCase for exported, camelCase for unexported

### Error Handling
- Return errors early from functions
- Use `fmt.Errorf` to wrap errors with context
- Check and handle all errors appropriately
- Use `defer` for resource cleanup (database connections, file handles)

### Types and Structs
- Define types close to where they're used
- Use meaningful field names
- Embed interfaces/types when appropriate

### General
- Use `context.Context` for cancellable operations
- Prefer explicit error checking over panic/recover
- Keep functions focused on single responsibilities
- Use meaningful variable names (avoid single-letter vars except in loops)