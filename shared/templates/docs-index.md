# {{.Name}} Documentation

{{.Repository.Description}}

## Quick Navigation

### 🚀 [Getting Started](getting-started.md)
Everything you need to get up and running with {{.Name}}.

### 📦 [Packages](packages/README.md)
Package overviews, installation, and quick start guides.

### 📚 [API Reference](api-reference/README.md)
Complete API documentation for all packages.

### 📖 [Examples](examples/README.md)
Working examples and tutorials.

### 📘 [Guides](guides/README.md)
In-depth guides and best practices.

## Package Overview

{{- range .Packages}}
### {{.Name}}

{{.Description}}

- [Package Overview](packages/{{.Name}}/README.md)
- [API Reference](api-reference/{{.Name}}/api-reference.md)
- [Examples](examples/{{.Name}}/README.md)
- [Best Practices](guides/{{.Name}}/best-practices.md)
{{- end}}

## External Resources

- [GitHub Repository](https://github.com/{{.Owner}}/{{.Name}})
- [pkg.go.dev Documentation](https://pkg.go.dev/{{.ImportPath}})
- [Issues & Support](https://github.com/{{.Owner}}/{{.Name}}/issues)

## Contributing

See our [Contributing Guide](guides/contributing.md) to get started.