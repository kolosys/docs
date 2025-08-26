package main

import (
	"encoding/json"
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

// Config represents the documentation configuration
type Config struct {
	Repository RepositoryConfig `json:"repository"`
	Packages   []PackageConfig  `json:"packages"`
	Docs       DocsConfig       `json:"docs"`
	Discovery  DiscoveryConfig  `json:"discovery"`
	Output     OutputConfig     `json:"output"`
}

type RepositoryConfig struct {
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	Description string `json:"description"`
	ImportPath  string `json:"import_path"`
}

type PackageConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	Path        string `json:"path,omitempty"` // For monorepos
}

type DocsConfig struct {
	RootDir      string `json:"root_dir"`
	DocsDir      string `json:"docs_dir"`
	TemplatesDir string `json:"templates_dir"`
}

type DiscoveryConfig struct {
	Enabled              bool     `json:"enabled"`
	ExcludePatterns      []string `json:"exclude_patterns"`
	IncludeOnlyWithGodoc bool     `json:"include_only_with_godoc"`
}

type OutputConfig struct {
	GenerateCombinedAPI bool `json:"generate_combined_api"`
	GenerateExamples    bool `json:"generate_examples"`
	Verbose             bool `json:"verbose"`
}

// DocGenerator generates GitBook documentation from Go packages
type DocGenerator struct {
	config Config
	fset   *token.FileSet
}

// PackageDoc represents documentation for a package
type PackageDoc struct {
	Name        string
	ImportPath  string
	Doc         string
	Functions   []FunctionDoc
	Types       []TypeDoc
	Constants   []ValueDoc
	Variables   []ValueDoc
	Examples    []ExampleDoc
}

// FunctionDoc represents a function's documentation
type FunctionDoc struct {
	Name      string
	Doc       string
	Signature string
	Examples  []ExampleDoc
}

// TypeDoc represents a type's documentation
type TypeDoc struct {
	Name       string
	Doc        string
	Decl       string
	Kind       string        // "struct", "interface", "type", etc.
	Fields     []FieldDoc    // For structs
	Methods    []FunctionDoc
	Examples   []ExampleDoc
	Underlying string        // For type aliases
}

// FieldDoc represents a struct field
type FieldDoc struct {
	Name string
	Type string
	Tag  string
	Doc  string
}

// ValueDoc represents a constant or variable
type ValueDoc struct {
	Name string
	Doc  string
	Decl string
}

// ExampleDoc represents an example
type ExampleDoc struct {
	Name string
	Code string
	Doc  string
}

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

	// Process shared templates if they exist
	err = generator.processSharedTemplates()
	if err != nil {
		fmt.Printf("❌ Error processing shared templates: %v\n", err)
		// Continue anyway - templates are optional
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

func loadConfig() (Config, error) {
	var config Config
	
	// Try multiple config file locations
	configPaths := []string{
		"docs-config.json",
		".docs-config.json",
		".config/docs.json",
		"kolosys-docs.json", // New centralized naming
	}
	
	var configFile string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configFile = path
			break
		}
	}
	
	if configFile == "" {
		return config, fmt.Errorf("no configuration file found. Please create docs-config.json")
	}
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		return config, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}
	
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}
	
	// Apply defaults for missing values
	if config.Docs.RootDir == "" {
		config.Docs.RootDir = "."
	}
	if config.Docs.DocsDir == "" {
		config.Docs.DocsDir = "docs"
	}
	if config.Repository.ImportPath == "" {
		config.Repository.ImportPath = fmt.Sprintf("github.com/%s/%s", 
			config.Repository.Owner, config.Repository.Name)
	}
	
	fmt.Printf("📄 Loaded configuration from %s\n", configFile)
	return config, nil
}

func (g *DocGenerator) GeneratePackageDocs(packageName string) error {
	// Parse the package
	pkgDoc, err := g.parsePackage(packageName)
	if err != nil {
		return fmt.Errorf("failed to parse package: %w", err)
	}

	// Create package directory
	pkgDir := filepath.Join(g.config.Docs.DocsDir, packageName)
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate README.md
	if err := g.generatePackageReadme(pkgDoc, pkgDir); err != nil {
		return fmt.Errorf("failed to generate README: %w", err)
	}

	// Generate API reference
	if err := g.generateAPIReference(pkgDoc, pkgDir); err != nil {
		return fmt.Errorf("failed to generate API reference: %w", err)
	}

	// Generate examples
	if err := g.generateExamples(pkgDoc, pkgDir); err != nil {
		return fmt.Errorf("failed to generate examples: %w", err)
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

func (g *DocGenerator) generatePackageReadme(pkg *PackageDoc, dir string) error {
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
- [{{ .Name }}](api-reference.md#{{ .Name | lower }}) - {{ .Doc | truncate }}
{{- end }}
{{- end }}

{{- if .Types }}
### Types  
{{- range .Types }}
- [{{ .Name }}](api-reference.md#{{ .Name | lower }}) - {{ .Doc | truncate }}
{{- end }}
{{- end }}

## Examples

See [examples](examples.md) for detailed usage examples.
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

	file, err := os.Create(filepath.Join(dir, "README.md"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, pkg)
}

func (g *DocGenerator) generateAPIReference(pkg *PackageDoc, dir string) error {
	tmpl := `# {{ .Name }} API Reference

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

	file, err := os.Create(filepath.Join(dir, "api-reference.md"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, pkg)
}

func (g *DocGenerator) processSharedTemplates() error {
	// Check if docs-templates directory exists
	templatesDir := "docs-templates"
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

	tmpl := `# {{ .Name }} Examples

## Basic Usage

` + "```go" + `
{{ .ExampleContent }}
` + "```" + `

## Advanced Examples

See the [examples directory]({{ .RepoURL }}/tree/main/examples/{{ .Name }}) for more comprehensive examples.
`

	t, err := template.New("examples").Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(dir, "examples.md"))
	if err != nil {
		return err
	}
	defer file.Close()

	data := struct {
		*PackageDoc
		ExampleContent string
		RepoURL        string
	}{
		PackageDoc:     pkg,
		ExampleContent: exampleContent,
		RepoURL:        fmt.Sprintf("https://github.com/%s/%s", g.config.Repository.Owner, g.config.Repository.Name),
	}

	return t.Execute(file, data)
}
