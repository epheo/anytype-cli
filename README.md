# Anytype CLI

A simple command-line interface for interacting with [Anytype](https://anytype.io/), enabling full management of spaces, objects, types, and more.

## Overview

Anytype CLI provides a complete set of commands for interacting with Anytype spaces, objects, types, lists, templates, and more. It uses the [anytype-go](https://github.com/epheo/anytype-go) SDK and spf13/cobra.

## Installation

### From Source

```bash
git clone https://github.com/epheo/anytype-cli.git
cd anytype-cli
go build -o bin/anytype-cli
```

### Requirements

- Go 1.18+
- Anytype app running locally with API enabled

## Getting Started

### Authentication

First, you need to authenticate with your local Anytype instance:

```bash
anytype-cli auth
```

This will prompt you to enter a verification code displayed in your Anytype app.

### Basic Usage

```bash
# Get CLI version information
anytype-cli version

# List all spaces
anytype-cli spaces list

# Search for objects
anytype-cli search --query "important" --space <space-id>

# Create a new page
anytype-cli objects create <space-id> --name "My New Page" --type "ot-page" --body "# Hello\n\nThis is my new page"
```

## Available Commands

### Global Options

- `--base-url`: Anytype API base URL (default: <http://localhost:31009>)
- `--config`: Custom config file location
- `--output`, `-o`: Output format (table, json, yaml)
- `--verbose`, `-v`: Enable verbose output

### Authentication Command

- `auth`: Authenticate with Anytype
  - `--force`: Force re-authentication

### Spaces

- `spaces list`: List all spaces
- `spaces get <space-id>`: Get details about a specific space
- `spaces create`: Create a new space
  - `--name`: Name for the space (required)
  - `--description`: Description for the space
  - `--icon`: Emoji icon for the space

### Objects

- `objects list <space-id>`: List objects in a space
- `objects get <space-id> <object-id>`: Get details about an object
- `objects create <space-id>`: Create a new object
  - `--name`: Name for the object (required)
  - `--type`: Type key for the object (default: ot-page)
  - `--description`: Description for the object
  - `--body`: Markdown body content
  - `--icon`: Emoji icon for the object
  - `--template`: Template ID to use
- `objects delete <space-id> <object-id>`: Delete an object
- `objects export <space-id> <object-id>`: Export an object in markdown format

### Types

- `types list <space-id>`: List all object types in a space
- `types get <space-id> <type-id>`: Get details about a specific object type
- `types templates <space-id> <type-id>`: List templates for a specific type
- `types template-get <space-id> <type-id> <template-id>`: Get details about a template

### Lists

- `lists views <space-id> <list-id>`: List views for a list
- `lists objects <space-id> <list-id> <view-id>`: List objects in a specific list view
- `lists add <space-id> <list-id> <object-id>...`: Add objects to a list
- `lists remove <space-id> <list-id> <object-id>`: Remove an object from a list

### Members

- `members list <space-id>`: List members in a space
- `members get <space-id> <member-id>`: Get details about a specific member

### Search

- `search`: Search for objects
  - `--query`: Search query string
  - `--types`: Filter by object types (comma-separated)
  - `--sort`: Property to sort by
  - `--direction`: Sort direction (asc or desc)
  - `--space`: Limit search to a specific space

## Examples

### Managing Spaces

```bash
# List all spaces
anytype-cli spaces list

# Get details about a space
anytype-cli spaces get <space-id>

# Create a new space
anytype-cli spaces create --name "Project Documentation" --description "Documentation for my projects" --icon "📚"
```

### Working with Objects

```bash
# List all objects in a space
anytype-cli objects list <space-id>

# Create a new page
anytype-cli objects create <space-id> --name "Meeting Notes" --type "ot-page" --body "# Meeting Notes\n\n## Agenda\n\n- Item 1\n- Item 2"

# Export an object as markdown
anytype-cli objects export <space-id> <object-id>
```

### Searching

```bash
# Search all spaces for objects containing "project"
anytype-cli search --query "project"

# Search in a specific space with filtering and sorting
anytype-cli search --query "task" --space <space-id> --types "ot-task" --sort "last_modified_date" --direction "desc"
```

### Working with Lists and Views

```bash
# List all views in a list
anytype-cli lists views <space-id> <list-id>

# List objects in a specific view
anytype-cli lists objects <space-id> <list-id> <view-id>

# Add an object to a list
anytype-cli lists add <space-id> <list-id> <object-id>
```

## License

Apache License 2.0
