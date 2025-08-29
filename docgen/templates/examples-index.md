# Examples

Complete working examples for {{.Name}}.

## Overview

This section contains practical examples demonstrating how to use {{.Name}} in real-world scenarios.

## Available Examples

{{- range .Packages}}

### {{.Name}}

{{.Description}}

- [Basic Usage]({{.Name}}/README.md)
- [Advanced Examples]({{.Name}}/advanced.md)
- [View Source Code](https://github.com/{{.Owner}}/{{$.Repository.Name}}/tree/main/examples/{{.Name}})

{{- end}}

## Running Examples

All examples can be run directly:

```bash
# Clone the repository
git clone https://github.com/{{.Owner}}/{{.Name}}.git
cd {{.Name}}

# Run a specific example
cd examples/[package-name]
go run main.go
```

## Contributing Examples

Have a great example? We'd love to include it! See our [Contributing Guide](../guides/contributing.md) for details.

## External Resources

- [GitHub Examples Directory](https://github.com/{{.Owner}}/{{.Name}}/tree/main/examples)
- [pkg.go.dev Examples](https://pkg.go.dev/{{.ImportPath}}#pkg-examples)
