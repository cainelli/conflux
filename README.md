# Conflux

A Terraform-like CLI for managing Confluence pages as code. Define pages as local `.md` files with YAML frontmatter and Confluence Storage Format XML body, then preview and publish with a `plan`/`apply` workflow.

## Setup

### Prerequisites

- A Confluence API token ([create one here](https://id.atlassian.com/manage-profile/security/api-tokens))

### Build

```bash
go build -o conflux ./cmd/conflux
```

### Configure Authentication

```bash
export CONFLUENCE_API_TOKEN=your_api_token_here
```

### Initialize a Project

```bash
conflux init
```

This creates `.conflux/config.yaml` and a `pages/` directory.

## Usage

### Preview changes

```bash
conflux plan
```

### Publish to Confluence

```bash
conflux apply
conflux apply --auto-approve  # skip confirmation
```

### Import existing pages from Confluence

```bash
conflux import --page-id 123456789
conflux import --page-id 123456789 --recursive  # include child pages
```

### Validate files

```bash
conflux validate
```

## Configuration

The `.conflux/config.yaml` file:

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

Files use YAML frontmatter for metadata and Confluence Storage Format XML for the body:

```markdown
---
title: "Page Title"
confluence:
  space_key: "MYSPACE"
parent: "123456789"
labels:
  - my-label
---

<h2>Heading</h2>
<p>Content in Confluence Storage Format...</p>
```

### Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `title` | Yes | Page title in Confluence |
| `confluence.space_key` | Yes | Confluence space key |
| `confluence.page_id` | No | Set automatically after first publish |
| `parent` | No | Parent page reference (relative path or page ID) |
| `labels` | No | Array of Confluence labels |
| `version` | No | Managed automatically |
| `last_sync` | No | Managed automatically |

## Project Structure

```
project-root/
├── .conflux/
│   └── config.yaml
├── pages/
│   ├── some-page.md
│   └── nested/
│       └── child-page.md
└── README.md
```

## License

[Apache 2.0](LICENSE)
