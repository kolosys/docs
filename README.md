# Kolosys Documentation System

Centralized documentation tools and templates for all Kolosys repositories.

## 🚀 Quick Setup

To add Kolosys documentation to any repository:

```bash
# Download and run the setup script
curl -sSL https://raw.githubusercontent.com/kolosys/docs/main/tools/setup-repo-docs.sh | bash -s your-repo-name

# Or manually:
curl -sSL https://raw.githubusercontent.com/kolosys/docs/main/tools/setup-repo-docs.sh -o setup-docs.sh
chmod +x setup-docs.sh
./setup-docs.sh your-repo-name
```

## 📁 Structure

```
kolosys/docs/
├── shared/
│   ├── scripts/           # Reusable documentation scripts
│   │   └── generate-docs.go
│   ├── templates/         # Shared markdown templates
│   │   ├── getting-started.md
│   │   └── installation.md
│   └── workflows/         # Reusable GitHub Actions
│       └── docs-workflow.yml
├── packages/              # Package-specific configurations
│   ├── ion/
│   └── timecapsule/
└── tools/                 # Setup and utility scripts
    └── setup-repo-docs.sh
```

## 🔧 Features

### ✅ Centralized Scripts
- **Single source of truth** for documentation generation
- **Automatic updates** when the shared scripts improve
- **Consistent output** across all repositories

### ✅ Shared Templates  
- **Common documentation patterns** (installation, getting started)
- **Branded formatting** consistent with Kolosys style
- **Easy customization** per repository

### ✅ Automated Workflows
- **Auto-generates** documentation from Go source code
- **Commits changes** automatically on push
- **Syncs with GitBook** for beautiful documentation sites

### ✅ Configuration-Driven
- **JSON configuration** for each repository
- **No external dependencies** - uses Go standard library
- **Flexible package discovery** for monorepos and single packages

## 📖 Usage

### For New Repositories

1. **Run the setup script:**
   ```bash
   curl -sSL https://raw.githubusercontent.com/kolosys/docs/main/tools/setup-repo-docs.sh | bash -s my-repo
   ```

2. **Customize configuration** (optional):
   ```json
   {
     "repository": {
       "name": "my-repo",
       "owner": "kolosys",
       "description": "My awesome Go package"
     },
     "packages": [
       {
         "name": "my-repo",
         "description": "Main package functionality",
         "priority": 1
       }
     ]
   }
   ```

3. **Commit and push:**
   ```bash
   git add .
   git commit -m "Add Kolosys documentation system"
   git push
   ```

4. **Documentation auto-generates!** 🎉

### For Existing Repositories

Update your repository to use the centralized system:

1. **Replace local scripts** with the setup script
2. **Update GitHub workflow** to use shared tools
3. **Configure GitBook** to sync with your repository

## 🔄 How It Works

1. **Repository pushes** trigger GitHub Actions
2. **Workflow downloads** latest documentation tools from this repo
3. **Generator parses** Go source code and extracts API information
4. **Templates render** markdown documentation
5. **Auto-commit** pushes generated docs back to repository
6. **GitBook syncs** and publishes beautiful documentation sites

## 🎯 Benefits

### For Developers
- **Zero maintenance** - documentation updates automatically
- **Consistent formatting** across all Kolosys projects
- **Rich API documentation** with full type information
- **Professional appearance** without manual work

### For Users
- **Comprehensive documentation** for every package
- **Searchable API reference** with Discord.js-style interface
- **Working examples** and getting started guides
- **Always up-to-date** with the latest code changes

## 📊 Supported Repositories

Current repositories using this system:

- **ion** - Concurrency primitives for Go
- **timecapsule** - Time-based data storage and retrieval

## 🛠️ Configuration

### Repository Configuration (`kolosys-docs.json`)

```json
{
  "repository": {
    "name": "package-name",
    "owner": "kolosys", 
    "description": "Package description"
  },
  "packages": [
    {
      "name": "package-name",
      "description": "Package description",
      "priority": 1,
      "path": "optional/subdirectory"
    }
  ],
  "docs": {
    "root_dir": ".",
    "docs_dir": "docs",
    "templates_dir": "docs-templates"
  },
  "output": {
    "generate_combined_api": true,
    "generate_examples": true,
    "verbose": true
  }
}
```

### GitBook Configuration (`.gitbook.yaml`)

```yaml
root: ./docs

structure:
  introduction: INTRODUCTION.md
  summary: SUMMARY.md
```

## 🚀 Advanced Usage

### Custom Templates

Override shared templates by creating local ones:

```bash
mkdir docs-templates
echo "# Custom Getting Started" > docs-templates/getting-started.md
```

### Multiple Packages (Monorepo)

Configure multiple packages in your config:

```json
{
  "packages": [
    {
      "name": "package1",
      "description": "First package",
      "priority": 1
    },
    {
      "name": "package2", 
      "description": "Second package",
      "priority": 2
    }
  ]
}
```

### Skip Auto-Commit

For CI/CD pipelines that handle commits differently:

```yaml
- name: Generate docs only
  run: |
    go run .kolosys-docs/generate-docs.go
```

## 🤝 Contributing

To improve the shared documentation system:

1. **Fork this repository**
2. **Make changes** to shared scripts/templates
3. **Test with a sample repository**
4. **Submit a pull request**

Changes to shared scripts automatically benefit all Kolosys repositories!

## 🔒 Branch Protection & Security

The documentation system supports repositories with GitHub branch protection rules and rulesets.

### **Pull Request Mode (Recommended)**

By default, the system creates pull requests instead of direct commits:

```yaml
jobs:
  generate-docs:
    uses: kolosys/docs/.github/workflows/docs-workflow.yml@main
    with:
      repository_name: "your-repo"
      create_pr: true  # Creates PRs instead of direct commits
    permissions:
      contents: write
      pull-requests: write
```

### **Direct Commit Mode (For Unprotected Branches)**

For repositories without branch protection:

```yaml
with:
  create_pr: false  # Direct commits to main branch
```

### **Personal Access Token Setup**

For additional security and bypassing some restrictions:

1. **Create a Personal Access Token** with these permissions:
   - `contents: write`
   - `pull-requests: write` 
   - `metadata: read`

2. **Add to repository secrets** as `DOCS_TOKEN`

3. **Token will be used automatically** by the workflow

### **GitHub Branch Protection Compatibility**

✅ **Works with all protection rules:**
- Required pull request reviews
- Required status checks  
- Restrict pushes to specific users/teams
- Repository rulesets
- Branch name patterns

✅ **Automated workflow:**
- Creates descriptive pull requests
- Includes detailed change summaries
- Auto-deletes feature branches after merge
- Safe to auto-merge documentation updates

## 📄 License

MIT License - see [LICENSE](LICENSE) file.