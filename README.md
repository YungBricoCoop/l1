# L1

`l1` is a cli tool made to simplify some common tasks

## Commands

```bash
l1 config --show-path
l1 config init
l1 config bkp
l1 config bkp-restore
l1 config --show
l1 config s3.url "https://s3.eu-west-1.amazonaws.com"
l1 config s3.default_bucket "mybucket"
l1 config ui.progress false

l1 push ./file.txt
l1 push ./file.txt --bucket other-bucket --key backups/file.txt
l1 pull ./file.txt
l1 pull ./downloads/file.txt --bucket other-bucket --key backups/file.txt
```

`l1` reads config from your user config directory by default:

- Linux: `~/.config/l1/config.toml`
- macOS: `~/Library/Application Support/l1/config.toml`
- Windows: `%APPDATA%\\l1\\config.toml`

You can override this for any command with `--config`:

```bash
l1 --config /tmp/l1.toml config --show
```

Example `config.toml`:

```toml
[s3]
url = "https://s3.eu-west-1.amazonaws.com"
region = "eu-west-1"
access_key = "env:AWS_ACCESS_KEY_ID"
secret_key = "env:AWS_SECRET_ACCESS_KEY"
default_bucket = "mybucket"

[ui]
color = true
progress = true
```

## Development tooling

This project uses the following tools in development and CI:

- [golangci-lint](https://github.com/golangci/golangci-lint) for formatting and linting Go code.
- [pre-commit](https://github.com/pre-commit/pre-commit) to run checks before commits.
- [REUSE](https://github.com/fsfe/reuse-tool) for license and copyright compliance.
- [bump-my-version](https://github.com/callowayproject/bump-my-version) for version bump automation.

## Todo

- [ ] Use Catpuccin colors
