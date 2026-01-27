# Run

Build and run gossm with the specified arguments.

## Arguments

- `<args>`: Optional. Arguments to pass to gossm (e.g., `start`, `list`, `exec`, `--help`).

## Instructions

Build and run the gossm binary with the provided arguments:

```bash
go run . <args>
```

Or if no arguments provided, show the help:

```bash
go run . --help
```

Report the output to the user.

## Available Commands

- `start` - Start an SSM session (use `-t` flag to specify instance ID)
- `list` - List all available SSM-connected instances
- `exec <instance-id> <command>` - Execute a command on a specific instance
