# Installation

## Requirements

- Go 1.22 or later
- No external dependencies required

## Install via go get

```bash
go get {{.ImportPath}}@latest
```

## Install specific version

```bash
go get {{.ImportPath}}@v0.1.0
```

## Verify installation

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

## Module integration

Add to your `go.mod`:

```bash
go mod init your-project
go get {{.ImportPath}}@latest
```

## Import packages

```go
import "{{.ImportPath}}"
```
