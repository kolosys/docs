package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	generator := &DocGenerator{
		config: config,
	}

	if config.Output.Verbose {
		fmt.Printf("🚀 Generating documentation for %s/%s (%d packages)...\n",
			config.Repository.Owner, config.Repository.Name, len(config.Packages))
	}

	// Create docs directory structure
	if err := generator.createDocumentationStructure(); err != nil {
		fmt.Printf("❌ Failed to create docs directory structure: %v\n", err)
		os.Exit(1)
	}

	// Process shared templates if they exist
	err = generator.processSharedTemplates()
	if err != nil {
		fmt.Printf("❌ Error processing shared templates: %v\n", err)
		// Continue anyway - templates are optional
	}

	// Generate structured documentation indexes
	if err := generator.generateDocumentationIndexes(); err != nil {
		fmt.Printf("❌ Error generating documentation indexes: %v\n", err)
		// Continue anyway - indexes are optional but recommended
	}

	// Copy repository README.md to docs directory
	if err := generator.copyRepositoryReadme(); err != nil {
		fmt.Printf("❌ Error copying repository README: %v\n", err)
		// Continue anyway - README copy is optional
	}

	for _, pkg := range config.Packages {
		if config.Output.Verbose {
			fmt.Printf("📝 Generating documentation for %s...\n", pkg.Name)
		}
		err := generator.GeneratePackageDocs(pkg.Name)
		if err != nil {
			fmt.Printf("❌ Error generating docs for %s: %v\n", pkg.Name, err)
			continue
		}
		if config.Output.Verbose {
			fmt.Printf("✅ Generated documentation for %s\n", pkg.Name)
		}
	}

	fmt.Printf("🎉 Documentation generation complete for %s!\n", config.Repository.Name)
}

func (g *DocGenerator) GeneratePackageDocs(packageName string) error {
	// Parse the package
	pkgDoc, err := g.parsePackage(packageName)
	if err != nil {
		return fmt.Errorf("failed to parse package: %w", err)
	}

	// Generate package overview as single file in packages directory
	packagesDir := filepath.Join(g.config.Docs.DocsDir, "packages")
	if err := os.MkdirAll(packagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create packages directory: %w", err)
	}

	// Generate package README as packages/packagename.md
	packageFile := filepath.Join(packagesDir, packageName+".md")
	if err := g.generatePackageMarkdown(pkgDoc, packageFile); err != nil {
		return fmt.Errorf("failed to generate package markdown: %w", err)
	}

	// Generate API reference as single file in api-reference directory
	apiDir := filepath.Join(g.config.Docs.DocsDir, "api-reference")
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		return fmt.Errorf("failed to create API directory: %w", err)
	}

	// Generate detailed API reference as api-reference/packagename.md
	apiFile := filepath.Join(apiDir, packageName+".md")
	if err := g.generateAPIMarkdown(pkgDoc, apiFile); err != nil {
		return fmt.Errorf("failed to generate API reference: %w", err)
	}

	// Generate examples in examples directory
	examplesDir := filepath.Join(g.config.Docs.DocsDir, "examples", packageName)
	if err := os.MkdirAll(examplesDir, 0755); err != nil {
		return fmt.Errorf("failed to create examples directory: %w", err)
	}

	if err := g.generateExamples(pkgDoc, examplesDir); err != nil {
		return fmt.Errorf("failed to generate examples: %w", err)
	}

	// Generate getting-started file for package in getting-started directory
	gettingStartedDir := filepath.Join(g.config.Docs.DocsDir, "getting-started")
	if err := os.MkdirAll(gettingStartedDir, 0755); err != nil {
		return fmt.Errorf("failed to create getting-started directory: %w", err)
	}

	gettingStartedFile := filepath.Join(gettingStartedDir, packageName+".md")
	if err := g.generateGettingStartedMarkdown(pkgDoc, gettingStartedFile); err != nil {
		return fmt.Errorf("failed to generate getting-started file: %w", err)
	}

	// Generate guides in package-specific directory structure
	packageGuidesDir := filepath.Join(g.config.Docs.DocsDir, "guides", packageName)
	if err := os.MkdirAll(packageGuidesDir, 0755); err != nil {
		return fmt.Errorf("failed to create package guides directory: %w", err)
	}

	if err := g.generatePackageGuides(pkgDoc, packageGuidesDir); err != nil {
		return fmt.Errorf("failed to generate guides: %w", err)
	}

	return nil
}

func (g *DocGenerator) parsePackage(packageName string) (*PackageDoc, error) {
	// Find package path - could be in root or subdirectory
	var pkgPath string

	// Try different possible locations
	possiblePaths := []string{
		filepath.Join(g.config.Docs.RootDir, packageName),
		filepath.Join(g.config.Docs.RootDir), // For single-package repos
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			pkgPath = path
			break
		}
	}

	if pkgPath == "" {
		return nil, fmt.Errorf("package directory not found for %s", packageName)
	}

	fset := token.NewFileSet()

	// Create a filter function to exclude _test.go files
	filter := func(info os.FileInfo) bool {
		// Exclude _test.go files
		if strings.HasSuffix(info.Name(), "_test.go") {
			return false
		}
		// Include all other .go files
		return strings.HasSuffix(info.Name(), ".go")
	}

	pkgs, err := parser.ParseDir(fset, pkgPath, filter, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var pkg *ast.Package
	for _, p := range pkgs {
		if !strings.HasSuffix(p.Name, "_test") {
			pkg = p
			break
		}
	}

	if pkg == nil {
		return nil, fmt.Errorf("no package found in %s", pkgPath)
	}

	docPkg := doc.New(pkg, "./"+packageName, doc.AllDecls)

	// Convert to our structure
	pkgDoc := &PackageDoc{
		Name:       docPkg.Name,
		ImportPath: g.config.Repository.ImportPath + "/" + packageName,
		Doc:        docPkg.Doc,
	}

	// Store fset for detailed type extraction
	g.fset = fset

	// Extract functions
	for _, f := range docPkg.Funcs {
		// Skip test functions
		if strings.HasPrefix(f.Name, "Test") || strings.HasPrefix(f.Name, "Benchmark") {
			continue
		}

		pkgDoc.Functions = append(pkgDoc.Functions, FunctionDoc{
			Name:      f.Name,
			Doc:       f.Doc,
			Signature: g.getFunctionSignature(f),
		})
	}

	// Extract types
	for _, t := range docPkg.Types {
		typeDoc := TypeDoc{
			Name:       t.Name,
			Doc:        t.Doc,
			Decl:       g.getTypeDecl(t),
			Kind:       g.getTypeKind(t),
			Fields:     g.getTypeFields(t),
			Underlying: g.getTypeUnderlying(t),
		}

		// Extract methods
		for _, m := range t.Methods {
			typeDoc.Methods = append(typeDoc.Methods, FunctionDoc{
				Name:      m.Name,
				Doc:       m.Doc,
				Signature: g.getFunctionSignature(m),
			})
		}

		pkgDoc.Types = append(pkgDoc.Types, typeDoc)
	}

	// Extract constants
	for _, c := range docPkg.Consts {
		for _, spec := range c.Decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range valueSpec.Names {
					pkgDoc.Constants = append(pkgDoc.Constants, ValueDoc{
						Name: name.Name,
						Doc:  c.Doc,
						Decl: g.getValueDecl(valueSpec),
					})
				}
			}
		}
	}

	// Extract variables
	for _, v := range docPkg.Vars {
		for _, spec := range v.Decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range valueSpec.Names {
					pkgDoc.Variables = append(pkgDoc.Variables, ValueDoc{
						Name: name.Name,
						Doc:  v.Doc,
						Decl: g.getValueDecl(valueSpec),
					})
				}
			}
		}
	}

	return pkgDoc, nil
}

// Helper methods for type extraction...
func (g *DocGenerator) getFunctionSignature(f *doc.Func) string {
	if f.Decl != nil {
		var buf strings.Builder
		err := format.Node(&buf, g.fset, f.Decl)
		if err == nil {
			return buf.String()
		}
	}
	return fmt.Sprintf("func %s(...)", f.Name)
}

func (g *DocGenerator) getTypeDecl(t *doc.Type) string {
	if t.Decl != nil {
		var buf strings.Builder
		err := format.Node(&buf, g.fset, t.Decl)
		if err == nil {
			return buf.String()
		}
	}
	return fmt.Sprintf("type %s", t.Name)
}

func (g *DocGenerator) getTypeKind(t *doc.Type) string {
	if t.Decl != nil && len(t.Decl.Specs) > 0 {
		if typeSpec, ok := t.Decl.Specs[0].(*ast.TypeSpec); ok {
			switch typeSpec.Type.(type) {
			case *ast.StructType:
				return "struct"
			case *ast.InterfaceType:
				return "interface"
			case *ast.FuncType:
				return "function"
			case *ast.ArrayType:
				return "array"
			case *ast.MapType:
				return "map"
			case *ast.ChanType:
				return "channel"
			default:
				return "type"
			}
		}
	}
	return "type"
}

func (g *DocGenerator) getTypeFields(t *doc.Type) []FieldDoc {
	var fields []FieldDoc

	if t.Decl != nil && len(t.Decl.Specs) > 0 {
		if typeSpec, ok := t.Decl.Specs[0].(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				for _, field := range structType.Fields.List {
					fieldDoc := FieldDoc{}

					// Get field name
					if len(field.Names) > 0 {
						fieldDoc.Name = field.Names[0].Name
					} else {
						// Embedded field
						if ident, ok := field.Type.(*ast.Ident); ok {
							fieldDoc.Name = ident.Name
						}
					}

					// Get field type
					var buf strings.Builder
					err := format.Node(&buf, g.fset, field.Type)
					if err == nil {
						fieldDoc.Type = buf.String()
					}

					// Get field tag
					if field.Tag != nil {
						fieldDoc.Tag = field.Tag.Value
					}

					// Get field documentation
					if field.Doc != nil {
						fieldDoc.Doc = strings.TrimSpace(field.Doc.Text())
					} else if field.Comment != nil {
						fieldDoc.Doc = strings.TrimSpace(field.Comment.Text())
					}

					fields = append(fields, fieldDoc)
				}
			}
		}
	}

	return fields
}

func (g *DocGenerator) getTypeUnderlying(t *doc.Type) string {
	if t.Decl != nil && len(t.Decl.Specs) > 0 {
		if typeSpec, ok := t.Decl.Specs[0].(*ast.TypeSpec); ok {
			if _, ok := typeSpec.Type.(*ast.StructType); !ok {
				if _, ok := typeSpec.Type.(*ast.InterfaceType); !ok {
					// This is a type alias or named type
					var buf strings.Builder
					err := format.Node(&buf, g.fset, typeSpec.Type)
					if err == nil {
						return buf.String()
					}
				}
			}
		}
	}
	return ""
}

func (g *DocGenerator) getValueDecl(spec *ast.ValueSpec) string {
	var buf strings.Builder
	err := format.Node(&buf, g.fset, spec)
	if err == nil {
		return buf.String()
	}

	if len(spec.Names) > 0 {
		return spec.Names[0].Name
	}
	return ""
}

func (g *DocGenerator) generatePackageMarkdown(pkg *PackageDoc, filePath string) error {
	tmpl := `# {{ .Name }}

{{ .Doc }}

## Installation

` + "```bash" + `
go get {{ .ImportPath }}
` + "```" + `

## Quick Start

` + "```go" + `
package main

import "{{ .ImportPath }}"

func main() {
    // Your code here
}
` + "```" + `

## API Reference

{{- if .Functions }}
### Functions
{{- range .Functions }}
- [{{ .Name }}](../api-reference/{{ $.Name }}.md#{{ .Name | lower }}) - {{ .Doc | truncate }}
{{- end }}
{{- end }}

{{- if .Types }}
### Types  
{{- range .Types }}
- [{{ .Name }}](../api-reference/{{ $.Name }}.md#{{ .Name | lower }}) - {{ .Doc | truncate }}
{{- end }}
{{- end }}

## Examples

See [examples](../examples/{{ .Name }}/README.md) for detailed usage examples.

## Resources

- [API Reference](../api-reference/{{ .Name }}.md) - Complete API documentation
- [Examples](../examples/{{ .Name }}/README.md) - Working examples
- [Best Practices](../guides/{{ .Name }}/best-practices.md) - Recommended patterns
`

	funcMap := template.FuncMap{
		"lower": strings.ToLower,
		"truncate": func(s string) string {
			if len(s) > 100 {
				return s[:97] + "..."
			}
			return s
		},
	}

	t, err := template.New("readme").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, pkg)
}

func (g *DocGenerator) generateAPIMarkdown(pkg *PackageDoc, filePath string) error {
	tmpl := `# {{ .Name }} API

{{- if .Functions }}
## Functions

{{- range .Functions }}
### {{ .Name }}

{{ .Doc }}

` + "```go" + `
{{ .Signature }}
` + "```" + `

{{- end }}
{{- end }}

{{- if .Types }}
## Types

{{- range .Types }}
### {{ .Name }}

{{ .Doc }}

` + "```go" + `
{{ .Decl }}
` + "```" + `

{{- if eq .Kind "struct" }}
{{- if .Fields }}
#### Fields

| Field | Type | Description |
|-------|------|-------------|
{{- range .Fields }}
| ` + "`{{ .Name }}`" + ` | ` + "`{{ .Type }}`" + ` | {{ .Doc | oneline }} |
{{- end }}
{{- end }}
{{- end }}

{{- if .Underlying }}
#### Underlying Type

` + "```go" + `
{{ .Underlying }}
` + "```" + `
{{- end }}

{{- if .Methods }}
#### Methods

{{- range .Methods }}
##### {{ .Name }}

{{ .Doc }}

` + "```go" + `
{{ .Signature }}
` + "```" + `

{{- end }}
{{- end }}
{{- end }}
{{- end }}

{{- if .Constants }}
## Constants

{{- range .Constants }}
### {{ .Name }}

{{ .Doc }}

` + "```go" + `
{{ .Decl }}
` + "```" + `

{{- end }}
{{- end }}

{{- if .Variables }}
## Variables

{{- range .Variables }}
### {{ .Name }}

{{ .Doc }}

` + "```go" + `
{{ .Decl }}
` + "```" + `

{{- end }}
{{- end }}
`

	funcMap := template.FuncMap{
		"oneline": func(s string) string {
			// Convert to single line and trim
			lines := strings.Split(strings.TrimSpace(s), "\n")
			if len(lines) > 0 {
				return strings.TrimSpace(lines[0])
			}
			return ""
		},
	}

	t, err := template.New("api").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, pkg)
}

func (g *DocGenerator) generateGettingStartedMarkdown(pkg *PackageDoc, filePath string) error {
	tmpl := `# Getting Started with {{ .Name }}

{{ .Doc }}

## Installation

` + "```bash" + `
go get {{ .ImportPath }}
` + "```" + `

## Quick Start

` + "```go" + `
package main

import "{{ .ImportPath }}"

func main() {
    // Your code here
    fmt.Println("Hello from {{ .Name }}!")
}
` + "```" + `

## Basic Usage

{{- if .Functions }}
### Functions
{{- range .Functions }}
- **{{ .Name }}** - {{ .Doc | oneline }}
{{- end }}
{{- end }}

{{- if .Types }}
### Types  
{{- range .Types }}
- **{{ .Name }}** - {{ .Doc | oneline }}
{{- end }}
{{- end }}

## Next Steps

- [Package Overview](../packages/{{ .Name }}.md) - Complete package information
- [API Reference](../api-reference/{{ .Name }}.md) - Detailed API documentation
- [Examples](../examples/{{ .Name }}/README.md) - Working examples and tutorials  
- [Best Practices](../guides/{{ .Name }}/best-practices.md) - Recommended usage patterns
- [Common Patterns](../guides/{{ .Name }}/patterns.md) - Common implementation patterns
`

	funcMap := template.FuncMap{
		"oneline": func(s string) string {
			// Convert to single line and trim
			lines := strings.Split(strings.TrimSpace(s), "\n")
			if len(lines) > 0 {
				return strings.TrimSpace(lines[0])
			}
			return ""
		},
	}

	t, err := template.New("getting-started").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, pkg)
}

func (g *DocGenerator) processSharedTemplates() error {
	// Check if templates directory exists
	templatesDir := "templates"
	if g.config.Docs.TemplatesDir != "" {
		templatesDir = g.config.Docs.TemplatesDir
	}

	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		// No templates directory, skip processing
		return nil
	}

	// Create docs directory if it doesn't exist
	docsDir := g.config.Docs.DocsDir
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	// Template data for shared templates
	templateData := struct {
		Repository RepositoryConfig
		ImportPath string
		Owner      string
		Name       string
	}{
		Repository: g.config.Repository,
		ImportPath: g.config.Repository.ImportPath,
		Owner:      g.config.Repository.Owner,
		Name:       g.config.Repository.Name,
	}

	// Process each template file
	templateFiles, err := filepath.Glob(filepath.Join(templatesDir, "*.md"))
	if err != nil {
		return fmt.Errorf("failed to find template files: %w", err)
	}

	for _, templateFile := range templateFiles {
		// Read template file
		templateContent, err := os.ReadFile(templateFile)
		if err != nil {
			fmt.Printf("⚠️  Failed to read template %s: %v\n", templateFile, err)
			continue
		}

		// Parse and execute template
		tmpl, err := template.New(filepath.Base(templateFile)).Parse(string(templateContent))
		if err != nil {
			fmt.Printf("⚠️  Failed to parse template %s: %v\n", templateFile, err)
			continue
		}

		// Create output file
		outputFile := filepath.Join(docsDir, filepath.Base(templateFile))
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Printf("⚠️  Failed to create output file %s: %v\n", outputFile, err)
			continue
		}

		// Execute template
		err = tmpl.Execute(file, templateData)
		file.Close()

		if err != nil {
			fmt.Printf("⚠️  Failed to execute template %s: %v\n", templateFile, err)
			continue
		}

		if g.config.Output.Verbose {
			fmt.Printf("📄 Processed template: %s -> %s\n", templateFile, outputFile)
		}
	}

	return nil
}

func (g *DocGenerator) generateExamples(pkg *PackageDoc, dir string) error {
	// Try to read examples from the examples directory
	exampleDir := filepath.Join("examples", pkg.Name)
	exampleFile := filepath.Join(exampleDir, "main.go")

	var exampleContent string
	if content, err := os.ReadFile(exampleFile); err == nil {
		exampleContent = string(content)
	} else {
		// Try single-package examples
		if content, err := os.ReadFile("example_test.go"); err == nil {
			exampleContent = string(content)
		} else {
			exampleContent = fmt.Sprintf(`package main

import (
    "context"
    "fmt"
    "log"
    
    "%s"
)

func main() {
    // Example usage of %s
    fmt.Println("See package documentation for examples")
}`, pkg.ImportPath, pkg.Name)
		}
	}

	// Generate main README for the package examples
	readmeContent := fmt.Sprintf(`# %s Examples

## Overview

This section contains working examples demonstrating how to use %s effectively.

## Available Examples

### Basic Usage

- [Basic Operations](basic.md) - Getting started with %s
- [Configuration](configuration.md) - Configuration options and patterns

### Advanced Usage

- [Advanced Patterns](advanced.md) - Complex usage scenarios
- [Integration](integration.md) - Integration with other packages
- [Performance](performance.md) - Performance optimization techniques

## Running Examples

To run the examples:

`+"```bash"+`
# Clone the repository
git clone https://github.com/%s/%s.git
cd %s

# Run the basic example
cd examples/%s
go run main.go
`+"```"+`

## Source Code

`+"```go"+`
%s
`+"```"+`

## More Examples

See the [GitHub examples directory](https://github.com/%s/%s/tree/main/examples/%s) for more comprehensive examples.
`, pkg.Name, pkg.Name, pkg.Name, g.config.Repository.Owner, g.config.Repository.Name, g.config.Repository.Name, pkg.Name, exampleContent, g.config.Repository.Owner, g.config.Repository.Name, pkg.Name)

	// Write the main README
	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to write examples README: %w", err)
	}

	// Generate basic example files
	examples := map[string]string{
		"basic.md": fmt.Sprintf(`# %s Basic Examples

## Getting Started

Basic usage patterns for %s.

### Example 1: Simple Usage

`+"```go"+`
package main

import (
    "fmt"
    "%s"
)

func main() {
    // Basic usage example
    // TODO: Add actual basic usage
    fmt.Println("Basic %s example")
}
`+"```"+`

### Example 2: With Configuration

`+"```go"+`
package main

import (
    "fmt"
    "%s"
)

func main() {
    // Configuration example
    // TODO: Add actual configuration example
    fmt.Println("Configured %s example")
}
`+"```"+`
`, pkg.Name, pkg.Name, pkg.ImportPath, pkg.Name, pkg.ImportPath, pkg.Name),

		"advanced.md": fmt.Sprintf(`# %s Advanced Examples

## Complex Scenarios

Advanced usage patterns and integration examples.

### Example 1: Advanced Configuration

`+"```go"+`
package main

import (
    "context"
    "fmt"
    "%s"
)

func main() {
    // Advanced configuration example
    // TODO: Add actual advanced example
    fmt.Println("Advanced %s example")
}
`+"```"+`

### Example 2: Error Handling

`+"```go"+`
package main

import (
    "context"
    "fmt"
    "log"
    "%s"
)

func main() {
    // Error handling example
    // TODO: Add actual error handling example
    fmt.Println("Error handling %s example")
}
`+"```"+`
`, pkg.Name, pkg.ImportPath, pkg.Name, pkg.ImportPath, pkg.Name),
	}

	// Write example files
	for filename, content := range examples {
		filepath := filepath.Join(dir, filename)
		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write example %s: %w", filename, err)
		}
	}

	if g.config.Output.Verbose {
		fmt.Printf("📖 Generated examples for %s\n", pkg.Name)
	}

	return nil
}

func (g *DocGenerator) copyRepositoryReadme() error {
	// Try to find the repository README.md file
	readmePaths := []string{
		"README.md",
		"readme.md",
		"Readme.md",
		"README.MD",
	}

	var readmePath string
	for _, path := range readmePaths {
		if _, err := os.Stat(path); err == nil {
			readmePath = path
			break
		}
	}

	if readmePath == "" {
		if g.config.Output.Verbose {
			fmt.Printf("⚠️  No README.md found in repository root, skipping copy\n")
		}
		return nil
	}

	// Read the repository README
	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("failed to read repository README: %w", err)
	}

	// Copy to docs directory
	docsReadmePath := filepath.Join(g.config.Docs.DocsDir, "README.md")
	err = os.WriteFile(docsReadmePath, readmeContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write README to docs directory: %w", err)
	}

	if g.config.Output.Verbose {
		fmt.Printf("📄 Copied %s to %s\n", readmePath, docsReadmePath)
	}

	return nil
}

// createDocumentationStructure creates the organized directory structure
func (g *DocGenerator) createDocumentationStructure() error {
	dirs := []string{
		g.config.Docs.DocsDir,
		filepath.Join(g.config.Docs.DocsDir, "getting-started"),
		filepath.Join(g.config.Docs.DocsDir, "packages"),
		filepath.Join(g.config.Docs.DocsDir, "api-reference"),
		filepath.Join(g.config.Docs.DocsDir, "examples"),
		filepath.Join(g.config.Docs.DocsDir, "guides"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	if g.config.Output.Verbose {
		fmt.Printf("📁 Created documentation directory structure\n")
	}

	return nil
}

// generateDocumentationIndexes creates index files for each main section
func (g *DocGenerator) generateDocumentationIndexes() error {
	// Template data for all indexes - include everything from Config
	templateData := struct {
		Repository RepositoryConfig
		Packages   []PackageConfig
		ImportPath string
		Owner      string
		Name       string
		Config     Config
	}{
		Repository: g.config.Repository,
		Packages:   g.config.Packages,
		ImportPath: g.config.Repository.ImportPath,
		Owner:      g.config.Repository.Owner,
		Name:       g.config.Repository.Name,
		Config:     g.config,
	}

	// Generate Getting Started index
	if err := g.generateTemplateFile("getting-started.md",
		filepath.Join(g.config.Docs.DocsDir, "getting-started", "README.md"), templateData); err != nil {
		return fmt.Errorf("failed to generate getting-started index: %w", err)
	}

	// Generate Packages index
	if err := g.generateTemplateFile("packages-index.md",
		filepath.Join(g.config.Docs.DocsDir, "packages", "README.md"), templateData); err != nil {
		return fmt.Errorf("failed to generate packages index: %w", err)
	}

	// Generate API Reference index
	if err := g.generateTemplateFile("api-reference-index.md",
		filepath.Join(g.config.Docs.DocsDir, "api-reference", "README.md"), templateData); err != nil {
		return fmt.Errorf("failed to generate API reference index: %w", err)
	}

	// Generate Examples index
	if err := g.generateTemplateFile("examples-index.md",
		filepath.Join(g.config.Docs.DocsDir, "examples", "README.md"), templateData); err != nil {
		return fmt.Errorf("failed to generate examples index: %w", err)
	}

	// Generate Guides index
	if err := g.generateTemplateFile("guides-index.md",
		filepath.Join(g.config.Docs.DocsDir, "guides", "README.md"), templateData); err != nil {
		return fmt.Errorf("failed to generate guides index: %w", err)
	}

	// Generate main docs index (README for docs directory)
	if err := g.generateTemplateFile("docs-index.md",
		filepath.Join(g.config.Docs.DocsDir, "README.md"), templateData); err != nil {
		return fmt.Errorf("failed to generate main docs index: %w", err)
	}

	// Generate GitBook configuration
	if err := g.generateGitBookConfig(); err != nil {
		return fmt.Errorf("failed to generate GitBook config: %w", err)
	}

	// Generate common guide files (only if they don't exist)
	if err := g.generatePreservableTemplateFile("contributing.md",
		filepath.Join(g.config.Docs.DocsDir, "guides", "contributing.md"), templateData); err != nil {
		// Not critical if this fails
		if g.config.Output.Verbose {
			fmt.Printf("⚠️  Failed to generate contributing guide: %v\n", err)
		}
	}

	if err := g.generatePreservableTemplateFile("faq.md",
		filepath.Join(g.config.Docs.DocsDir, "guides", "faq.md"), templateData); err != nil {
		// Not critical if this fails
		if g.config.Output.Verbose {
			fmt.Printf("⚠️  Failed to generate FAQ: %v\n", err)
		}
	}

	if g.config.Output.Verbose {
		fmt.Printf("📄 Generated documentation section indexes\n")
	}

	return nil
}

// generateTemplateFile processes a template and writes it to a file
func (g *DocGenerator) generateTemplateFile(templateName, outputPath string, data interface{}) error {
	// Try to find the template in templates directory
	templatesDir := "templates"
	if g.config.Docs.TemplatesDir != "" {
		templatesDir = g.config.Docs.TemplatesDir
	}

	templatePath := filepath.Join(templatesDir, templateName)

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		if g.config.Output.Verbose {
			fmt.Printf("⚠️  Template %s not found, skipping %s\n", templateName, outputPath)
		}
		return nil
	}

	// Read and process template
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	tmpl, err := template.New(templateName).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	// Create output file
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer file.Close()

	// Execute template
	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	if g.config.Output.Verbose {
		fmt.Printf("📄 Generated: %s\n", outputPath)
	}

	return nil
}

// generatePreservableTemplateFile processes a template and writes it to a file only if it doesn't exist
func (g *DocGenerator) generatePreservableTemplateFile(templateName, outputPath string, data interface{}) error {
	// Check if output file already exists - if so, skip it to preserve manual content
	if _, err := os.Stat(outputPath); err == nil {
		if g.config.Output.Verbose {
			fmt.Printf("📝 Skipped existing file: %s (preserving manual content)\n", filepath.Base(outputPath))
		}
		return nil
	}

	// File doesn't exist, so generate it using the normal template process
	return g.generateTemplateFile(templateName, outputPath, data)
}

// generatePackageGuides creates guide stubs for a package only if they don't exist
func (g *DocGenerator) generatePackageGuides(pkg *PackageDoc, dir string) error {
	// Create basic guide files for the package in directory structure
	guides := map[string]string{
		"best-practices.md": fmt.Sprintf(`# %s Best Practices

## Overview

Best practices for using %s effectively.

## Performance Considerations

- Performance tips and optimization strategies
- Memory usage patterns
- Concurrency considerations

## Common Patterns

- Recommended usage patterns
- Anti-patterns to avoid
- Integration strategies

## Error Handling

- Error handling strategies
- Recovery patterns
- Debugging tips

<!-- 
GUIDE CONTENT NOTICE:
This file is only created if it doesn't exist. Once created, it won't be overwritten
by the documentation generator, so you can safely edit and maintain the content.
-->
`, pkg.Name, pkg.Name),

		"patterns.md": fmt.Sprintf(`# %s Common Patterns

## Overview

Common usage patterns and examples for %s.

## Basic Patterns

### Pattern 1: Basic Usage

`+"```go"+`
// Example basic usage pattern
// TODO: Add actual usage examples
`+"```"+`

## Advanced Patterns

### Pattern 1: Advanced Usage

`+"```go"+`
// Example advanced usage pattern  
// TODO: Add actual advanced examples
`+"```"+`

## Integration Patterns

- Integration with other packages
- Middleware patterns
- Testing patterns

<!-- 
GUIDE CONTENT NOTICE:
This file is only created if it doesn't exist. Once created, it won't be overwritten
by the documentation generator, so you can safely edit and maintain the content.
-->
`, pkg.Name, pkg.Name),

		"README.md": fmt.Sprintf(`# %s Guides

This directory contains guides and best practices for using the %s package.

## Available Guides

- [Best Practices](best-practices.md) - Recommended usage patterns and performance tips
- [Common Patterns](patterns.md) - Examples and implementation patterns

## Getting Help

If you need help with %s:
- Check the [API Reference](../../api-reference/%s.md)
- Review the [Examples](../../examples/%s/README.md)
- Visit the [Getting Started guide](../../getting-started/%s.md)
`, pkg.Name, pkg.Name, pkg.Name, pkg.Name, pkg.Name, pkg.Name),
	}

	for filename, content := range guides {
		filePath := filepath.Join(dir, filename)

		// Check if file already exists - if so, skip it to preserve manual content
		if _, err := os.Stat(filePath); err == nil {
			if g.config.Output.Verbose {
				fmt.Printf("📝 Skipped existing guide: %s (preserving manual content)\n", filename)
			}
			continue
		}

		// Only create the file if it doesn't exist
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write guide %s: %w", filename, err)
		}

		if g.config.Output.Verbose {
			fmt.Printf("📚 Created new guide stub: %s/%s\n", filepath.Base(dir), filename)
		}
	}

	return nil
}

// generateGitBookConfig creates .gitbook.yaml configuration file
func (g *DocGenerator) generateGitBookConfig() error {
	gitbookConfig := `root: ./docs

structure:
  readme: README.md
  summary: SUMMARY.md

redirects:
  previous/page: new-folder/page.md
`

	// Write .gitbook.yaml to project root (not docs directory)
	configPath := ".gitbook.yaml"
	if err := os.WriteFile(configPath, []byte(gitbookConfig), 0644); err != nil {
		return fmt.Errorf("failed to write .gitbook.yaml: %w", err)
	}

	if g.config.Output.Verbose {
		fmt.Printf("📄 Generated .gitbook.yaml configuration\n")
	}

	return nil
}
