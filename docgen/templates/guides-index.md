# Guides

In-depth guides and best practices for {{.Name}}.

## Getting Started

- [Installation & Setup](../getting-started.md)
- [Quick Start Guide](quick-start.md)
- [Basic Concepts](concepts.md)

## Best Practices

- [Performance Optimization](performance.md)
- [Error Handling](error-handling.md)
- [Testing Strategies](testing.md)
- [Production Deployment](deployment.md)

## Advanced Topics

- [Architecture Overview](architecture.md)
- [Extending {{.Name}}](extending.md)
- [Integration Patterns](integration.md)
- [Troubleshooting](troubleshooting.md)

## Package-Specific Guides

{{- range .Packages}}

### {{.Name}}

- [{{.Name}} Best Practices]({{.Name}}-best-practices.md)
- [Common Patterns]({{.Name}}-patterns.md)
{{- end}}

## Community Resources

- [Contributing Guide](contributing.md)
- [Code of Conduct](code-of-conduct.md)
- [Security Policy](security.md)
- [FAQ](faq.md)

## External Resources

- [GitHub Repository](https://github.com/{{.Owner}}/{{.Name}})
- [Discussions](https://github.com/{{.Owner}}/{{.Name}}/discussions)
- [Issues](https://github.com/{{.Owner}}/{{.Name}}/issues)
