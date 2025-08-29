# API Reference

Complete API documentation for {{.Name}}.

## Overview

{{.Repository.Description}}

## Packages

{{- range .Packages}}

- **[{{.Name}}]({{.Name}}/README.md)** - {{.Description}}
  {{- end}}

## Quick Navigation

{{- range .Packages}}

### {{.Name}}

- [Overview]({{.Name}}/README.md)
- [API Reference]({{.Name}}/api-reference.md)
- [Examples](../examples/{{.Name}}/README.md)
  {{- end}}

## External References

- [pkg.go.dev Documentation](https://pkg.go.dev/{{.ImportPath}})
- [GitHub Repository](https://github.com/{{.Owner}}/{{.Name}})
- [Examples Directory](https://github.com/{{.Owner}}/{{.Name}}/tree/main/examples)
