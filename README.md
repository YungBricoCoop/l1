# L1

`l1` is a cli tool made to simplify some common tasks

## Install

`l1` can be installed directly using the following one-liner commands:

Unix/macOS/Linux/Windows(Git Bash):

```bash
curl -sSfL https://raw.githubusercontent.com/YungBricoCoop/l1/main/install.sh | sh
```

Alpine:

```bash
wget -qO- https://raw.githubusercontent.com/YungBricoCoop/l1/main/install.sh | sh
```

Specific version:

```bash
curl -sSfL https://raw.githubusercontent.com/YungBricoCoop/l1/main/install.sh | sh -s -- v0.1.0
```

Custom install directory:

```bash
curl -sSfL https://raw.githubusercontent.com/YungBricoCoop/l1/main/install.sh | sh -s -- -b /usr/local/bin
```

Notes:

- On Unix, default install path is `/usr/local/bin` (fallback: `~/.local/bin`).
- On Windows Git Bash, default install path is `~/bin`.

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
l1 config ui.theme "catppuccin-mocha"
l1 config gi.templates "python macos visualstudiocode"

l1 gi
l1 gi m v p
l1 gi macos visualstudiocode python
l1 gi c
l1 gi list
l1 gi --list
l1 gi p,m,v
l1 gi p,m,v --output .gitignore.global

l1 push ./file.txt
l1 push ./file.txt --bucket other-bucket --key backups/file.txt
l1 pull ./file.txt
l1 pull ./downloads/file.txt --bucket other-bucket --key backups/file.txt
```

`l1 gi` shortcut mapping: `p=python`, `v=visualstudiocode`, `g=go`, `m=macos`, `r=rust`.

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
theme = "catppuccin-mocha"

[gi]
templates = ["python", "macos", "visualstudiocode"]
```

Supported `ui.theme` values:

- `catppuccin-mocha` (default)
- `catppuccin-frappe`
- `catppuccin-macchiato`
- `catppuccin-latte`

## Development tooling

This project uses the following tools in development and CI:

- [golangci-lint](https://github.com/golangci/golangci-lint) for formatting and linting Go code.
- [pre-commit](https://github.com/pre-commit/pre-commit) to run checks before commits.
- [REUSE](https://github.com/fsfe/reuse-tool) for license and copyright compliance.
- [bump-my-version](https://github.com/callowayproject/bump-my-version) for version bump automation.
