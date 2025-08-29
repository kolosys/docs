# Packages

Overview of all packages in {{.Name}}.

## Available Packages

{{- range .Packages}}
### [{{.Name}}]({{.Name}}.md)

{{.Description}}

- [Package Overview]({{.Name}}.md) - Installation, quick start, and overview
- [API Reference](../api-reference/{{.Name}}.md) - Detailed API documentation
- [Examples](../examples/{{.Name}}/README.md) - Working examples and tutorials
- [Best Practices](../guides/{{.Name}}-best-practices.md) - Recommended usage patterns

{{- end}}

## Quick Start

To get started with any package:

1. **Install**: `go get {{.ImportPath}}/[package-name]`
2. **Read Overview**: Check the package README for basic concepts
3. **Try Examples**: Run the examples to see it in action
4. **Check API**: Reference the detailed API documentation

## Package Architecture

{{.Name}} uses a modular architecture where each package can be used independently:

```
{{.ImportPath}}/
{{- range .Packages}}
├── {{.Name}}/     # {{.Description}}
{{- end}}
└── shared/       # Common utilities and interfaces
```

## External Resources

- [GitHub Repository](https://github.com/{{.Owner}}/{{.Name}})
- [pkg.go.dev Documentation](https://pkg.go.dev/{{.ImportPath}})
- [Examples Directory](https://github.com/{{.Owner}}/{{.Name}}/tree/main/examples)
