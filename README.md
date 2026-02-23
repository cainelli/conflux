# Conflux - Confluence as Code

Conflux is a Terraform-like CLI tool for managing Confluence Cloud pages as code. It enables declarative management of Confluence pages using local files with a familiar plan/apply workflow.

## Features

- 📝 **Declarative Management** - Define Confluence pages in local files with YAML frontmatter
- 🔄 **Plan/Apply Workflow** - Preview changes before applying them (like Terraform)
- 📥 **Import Existing Pages** - Import existing Confluence pages to local management
- 🏗️ **Page Hierarchies** - Manage parent-child page relationships
- 🏷️ **Label Management** - Automatically sync page labels
- ✅ **Validation** - Validate local page files before applying

## Installation

### From Source

```bash
go install github.com/getyourguide/disaster-recovery-plans/cmd/conflux@latest
```

### Build Locally

```bash
git clone https://github.com/getyourguide/disaster-recovery-plans.git
cd disaster-recovery-plans
go build -o conflux ./cmd/conflux
```

## Quick Start

### 1. Initialize a New Project

```bash
conflux init
```

This will create:
- `.conflux/config.yaml` - Configuration file
- `pages/` - Directory for your page files
- `pages/index.md` - Example page

### 2. Configure Authentication

Set your Confluence API token as an environment variable:

```bash
export CONFLUENCE_API_TOKEN=your_api_token_here
```

To create an API token, go to: https://id.atlassian.com/manage-profile/security/api-tokens

### 3. Create or Edit Pages

Create a new file in the `pages/` directory using Confluence Storage Format:

```markdown
---
title: "My First Page"
confluence:
  space_key: MYSPACE
labels:
  - documentation
---

<h1>My First Page</h1>

<p>This is my first page managed with Conflux!</p>

<h2>Features</h2>

<ul>
  <li>Easy to manage</li>
  <li>Version controlled</li>
  <li>Collaborative</li>
</ul>
```

### 4. Preview Changes

```bash
conflux plan
```

### 5. Apply Changes

```bash
conflux apply
```

## Configuration

The `.conflux/config.yaml` file contains your project configuration:

```yaml
confluence:
  base_url: https://your-domain.atlassian.net/wiki
  space_key: MYSPACE
  auth:
    type: token
    email: your-email@company.com
    token: ${CONFLUENCE_API_TOKEN}

project:
  pages_dir: pages

behavior:
  create_missing_parents: true
  update_labels: true
```

## Page File Format

Pages use YAML frontmatter for metadata and Confluence Storage Format for content:

```markdown
---
title: "Page Title"
confluence:
  page_id: "123456789"        # Set automatically after creation
  space_key: "MYSPACE"
parent: "runbooks/index"      # Relative path or page_id
labels:
  - runbook
  - database
version: 1
last_sync: "2026-02-03T10:30:00Z"
---

<h1>Page Content</h1>

<p>Your content here in Confluence Storage Format...</p>
```

### Frontmatter Fields

- **title** (required): The page title
- **confluence.space_key** (required): The Confluence space key
- **confluence.page_id**: The Confluence page ID (set automatically)
- **parent**: Parent page reference (relative path or page ID)
- **labels**: Array of labels to apply
- **version**: Page version number (managed automatically)
- **last_sync**: Last sync timestamp (managed automatically)

## Commands

### `conflux init`

Initialize a new Conflux project with interactive setup.

```bash
conflux init
```

Options:
- `--force, -f`: Overwrite existing configuration

### `conflux plan`

Preview changes that would be made to Confluence.

```bash
conflux plan
```

### `conflux apply`

Apply planned changes to Confluence.

```bash
conflux apply
```

Options:
- `--auto-approve, -y`: Skip confirmation prompt

### `conflux import`

Import existing Confluence pages to local management.

```bash
# Import a single page
conflux import --page-id 123456789

# Import a page and all its children
conflux import --page-id 123456789 --recursive
```

Options:
- `--page-id, -p`: Confluence page ID to import (required)
- `--recursive, -r`: Recursively import child pages

### `conflux validate`

Validate local page files.

```bash
conflux validate
```

### `conflux version`

Show version information.

```bash
conflux version
```

## Project Structure

```
project-root/
├── .conflux/
│   ├── config.yaml           # Configuration file
│   └── .gitignore
├── pages/
│   ├── runbooks/
│   │   ├── database-failover.md
│   │   └── service-restart.md
│   ├── architecture/
│   │   ├── overview.md
│   │   └── microservices/
│   │       └── auth-service.md
│   └── index.md
└── README.md
```

## Page Hierarchies

You can organize pages in a hierarchy using the `parent` field:

```markdown
---
title: "Child Page"
confluence:
  space_key: MYSPACE
parent: "parent-page"         # Relative path (without .md)
---
```

Or reference by page ID:

```markdown
---
title: "Child Page"
confluence:
  space_key: MYSPACE
parent: "123456789"           # Parent page ID
---
```

## Confluence Storage Format

Conflux uses Confluence Storage Format for page content. Common elements:

- **Headers**: `<h1>`, `<h2>`, etc.
- **Paragraphs**: `<p>Content</p>`
- **Bold/Italic**: `<strong>`, `<em>`
- **Lists**: `<ul>`, `<ol>`, `<li>`
- **Code blocks**: `<ac:structured-macro ac:name="code">...</ac:structured-macro>`
- **Links**: `<a href="url">text</a>`
- **Tables**: `<table>`, `<tr>`, `<td>`
- **Info panels**: `<ac:structured-macro ac:name="info">...</ac:structured-macro>`

For more details, see the [Confluence Storage Format documentation](https://confluence.atlassian.com/doc/confluence-storage-format-790796544.html).

## Workflow Examples

### Creating a New Runbook

```bash
# Create the file
cat > pages/runbooks/deployment.md <<EOF
---
title: "Deployment Runbook"
confluence:
  space_key: OPS
parent: "runbooks/index"
labels:
  - runbook
  - deployment
---

<h1>Deployment Runbook</h1>

<h2>Prerequisites</h2>

<ul>
  <li>Access to production environment</li>
  <li>Approval from team lead</li>
</ul>

<h2>Steps</h2>

<ol>
  <li>Check system health</li>
  <li>Create backup</li>
  <li>Deploy new version</li>
  <li>Verify deployment</li>
  <li>Monitor for issues</li>
</ol>
EOF

# Preview changes
conflux plan

# Apply to Confluence
conflux apply
```

### Updating an Existing Page

```bash
# Edit the file
vim pages/runbooks/deployment.md

# Preview changes
conflux plan

# Apply changes
conflux apply
```

### Importing Existing Documentation

```bash
# Import a documentation tree
conflux import --page-id 123456789 --recursive

# Review imported files
ls -R pages/

# Make changes and re-apply
conflux plan
conflux apply
```

## Best Practices

1. **Version Control**: Keep your pages directory in Git for version history
2. **Branch Strategy**: Use feature branches for major documentation changes
3. **Code Review**: Review changes via Pull Requests before applying
4. **Naming**: Use descriptive, lowercase filenames with hyphens
5. **Organization**: Group related pages in subdirectories
6. **Labels**: Use consistent labeling for easy discovery
7. **Backups**: Run `conflux import` periodically to backup content

## Troubleshooting

### "Config file not found"

Run `conflux init` in your project directory to create the configuration.

### "Authentication failed"

Ensure your API token is set correctly:

```bash
export CONFLUX_API_TOKEN=your_token
```

### "Page ID required for update"

The page hasn't been created yet. Run `conflux apply` to create it first.

### "Parent page not found"

Ensure the parent page exists and has been created/imported before creating child pages.

## Limitations (MVP)

The following features are not yet implemented:

- State management / drift detection
- Attachments
- Page comments
- Page restrictions / permissions
- Batch operations
- Parallel updates

These features may be added in future versions.

## Contributing

Contributions are welcome! Please open an issue or Pull Request.

## License

[Apache 2.0](LICENSE)

## Support

For issues and questions, please open a GitHub issue.
