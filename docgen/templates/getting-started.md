# Getting Started

This guide will help you get up and running quickly with {{.Name}}.

## Installation

### Requirements

- Go 1.21 or later
- No external dependencies required

### Install via go get

```bash
go get {{.ImportPath}}@latest
```

### Install specific version

```bash
go get {{.ImportPath}}@v0.1.0
```

### Verify installation

Create a simple test file:

```go
package main

import (
    "fmt"

    "{{.ImportPath}}"
)

func main() {
    fmt.Println("{{.Name}} installed successfully!")
}
```

Run it:

```bash
go run main.go
```

### Module integration

Add to your `go.mod`:

```bash
go mod init your-project
go get {{.ImportPath}}@latest
```

## Quick Start

Here's a simple example to get you started:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "{{.ImportPath}}"
)

func main() {
    // Your code here
    fmt.Println("Hello from {{.Name}}!")
}
```

## Available Packages

{{.Name}} provides the following packages:

{{- range .Packages}}

### [{{.Name}}](packages/{{.Name}}.md)

{{.Description}}

**Quick Links:**

- [Package Overview](packages/{{.Name}}.md) - Installation and getting started
- [API Reference](api-reference/{{.Name}}.md) - Complete API documentation
- [Examples](examples/{{.Name}}/README.md) - Working examples
- [Best Practices](guides/{{.Name}}-best-practices.md) - Recommended patterns

{{- end}}

## Next Steps

- [API Reference](api-reference/README.md) - Complete API documentation
- [Examples](examples/README.md) - Working examples and tutorials
- [Guides](guides/README.md) - In-depth guides and best practices
- [GitHub Repository](https://github.com/{{.Owner}}/{{.Name}})
