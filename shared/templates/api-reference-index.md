# API Reference

Complete API documentation for {{.Name}}.

## Overview

This section contains detailed API documentation for all packages. For package overviews and getting started guides, see the [Packages](../packages/README.md) section.

## Package APIs

{{- range .Packages}}
### [{{.Name}}]({{.Name}}/api-reference.md)

{{.Description}}

**[→ Full API Documentation]({{.Name}}/api-reference.md)**

Key APIs:
- Types and interfaces
- Functions and methods  
- Constants and variables
- Detailed usage examples

{{- end}}

## Navigation

- **[Packages](../packages/README.md)** - Package overviews and installation
- **[Examples](../examples/README.md)** - Working code examples
- **[Guides](../guides/README.md)** - Best practices and patterns

## External References

- [pkg.go.dev Documentation](https://pkg.go.dev/{{.ImportPath}}) - Go module documentation
- [GitHub Repository](https://github.com/{{.Owner}}/{{.Name}}) - Source code and issues