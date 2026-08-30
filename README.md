# Anytype CLI

Command line client for the [Anytype](https://anytype.io/) local API,
built on the [anytype-go](https://github.com/epheo/anytype-go) SDK.

## Install

```bash
go install github.com/epheo/anytype-cli@latest
```

Or from source:

```bash
git clone https://github.com/epheo/anytype-cli.git
cd anytype-cli
make build      # bin/anytype-cli, version stamped from git
make install    # /usr/local/bin/anytype-cli
```

Requires Go 1.23+ and a running Anytype app with the local API enabled.

## Authenticate

```bash
anytype-cli auth          # type the code shown by the app
anytype-cli auth status   # where credentials come from, and whether the key works
anytype-cli auth logout
```

The API key is stored in `~/.anytype-cli/config.yaml` (mode 0600).

Overrides:

- `--config <path>`: another config file
- `--base-url <url>` or `ANYTYPE_BASE_URL`: API address (default `http://localhost:31009`)
- `ANYTYPE_APP_KEY`: API key without a config file

## Choose a space

Most commands work inside one space. It is taken from, in order:

1. `--space` / `-s` on the command line
2. `ANYTYPE_SPACE` in the environment
3. `default_space` in the config file

```bash
anytype-cli config set default-space "My Space"
anytype-cli config get
anytype-cli config unset default-space
```

A space may be given by ID or by name.
Name matching is case-insensitive, exact first, then unique substring.

## Usage

```bash
anytype-cli spaces list
anytype-cli objects list --limit 20
anytype-cli objects get <object-id>
anytype-cli objects create --name "Notes" --type page --body "# Hello"
anytype-cli objects create --name "Notes" --body-file note.md
anytype-cli objects export <object-id> | sed 's/foo/bar/' | anytype-cli objects update <object-id> --body-file -
anytype-cli objects update <object-id> --property done=true --property due=2026-01-01
anytype-cli types get page
anytype-cli search "meeting" --types page,task --sort last_modified_date
anytype-cli search --filter 'name:contains:plan' --filter 'created_date:gt:2026-01-01'
anytype-cli search "meeting" --all-spaces
```

### Output

`-o table` (default), `-o json`, `-o yaml`.
JSON and YAML print the raw SDK structures; tables truncate long names but never IDs.

### Pagination

Every `list` command and `search` accept `--limit`, `--offset`, and `--all`.
Tables end with a footer showing how to fetch the next page.

### Commands

| Group | Commands | Space |
| --- | --- | --- |
| `spaces` | `list`, `get [space]`, `create`, `update [space]` | positional, falls back to current |
| `objects` | `list`, `get`, `create`, `update`, `delete`, `export` | current |
| `types` | `list`, `get`, `create`, `update`, `delete` | current |
| `templates` | `list <type>`, `get <type> <template-id>` | current |
| `properties` | `list`, `get`, `create`, `update`, `delete` | current |
| `tags` | `list <property-id>`, `get`, `create`, `update`, `delete` | current |
| `lists` | `views`, `objects`, `add`, `remove` | current |
| `members` | `list`, `get` | current |
| `search` | `[query]` | current, or all with `--all-spaces` |
| `auth` | `status`, `logout` | |
| `config` | `path`, `get`, `set`, `unset` | |
| `version`, `completion` | | |

Run `anytype-cli <group> <command> --help` for flags.

Notes:

- `<type>` arguments accept a key (`page`), a display name (`Page`), or an ID.
- `--property key=value` resolves the property format from the space, so the type is never stated.
  Lists are comma-separated; select and multi_select take tag IDs; an empty value clears the property.
- `--filter key:condition[:value]` works the same way. Conditions: eq, ne, in, nin, contains,
  ncontains, gt, lt, gte, lte, all, empty, nempty. `--match any` switches from and to or.
  Keys missing from the property list (such as `name`) are sent as text.
- `delete` archives; nothing is destroyed.
- Errors carry the API message and HTTP status, for example `Error: invalid api key (HTTP 401)`.

### Shell completion

```bash
source <(anytype-cli completion bash)
anytype-cli completion zsh > "${fpath[1]}/_anytype-cli"
anytype-cli completion fish > ~/.config/fish/completions/anytype-cli.fish
```

`--space` and the `spaces get` argument complete by ID and name.

## Development

```bash
make test
make lint
```

## License

Apache License 2.0
