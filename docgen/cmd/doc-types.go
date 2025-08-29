package main

import "go/token"

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
	Name       string
	ImportPath string
	Doc        string
	Functions  []FunctionDoc
	Types      []TypeDoc
	Constants  []ValueDoc
	Variables  []ValueDoc
	Examples   []ExampleDoc
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
	Kind       string     // "struct", "interface", "type", etc.
	Fields     []FieldDoc // For structs
	Methods    []FunctionDoc
	Examples   []ExampleDoc
	Underlying string // For type aliases
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
