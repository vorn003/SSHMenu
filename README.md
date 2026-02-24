# SSHMenu

A terminal-based session manager for SSH connections, supporting two-step navigation and search.

## Features
- Two-step UI: Select project, then server (when no filter is used)
- Flat filtered list: Search servers by name or description using a command-line parameter
- Inplace update: Reloads YAML config on each action
- Custom SSH commands per server or global command template

## Usage
### Installation

1. **Download the latest release:**
   - Go to the [GitHub Releases page](https://github.com/vorn003/session-manager/releases)
   - Download the `sshmenu` binary for your platform (e.g., `sshmenu-linux-amd64`, etc.)
   - Make it executable if needed:
     ```
     chmod +x sshmenu
     ```
   - (Optional) Move the binary to your PATH:
     ```
     sudo mv sshmenu /usr/local/bin/
     ```

2. **Configuration file location:**
   - By default, SSHMenu will look for the configuration file at `~/.config/sshmenu/sshmenu.yaml`.
   - If that file does not exist, it will use `sshmenu.yaml` in the same directory as the binary.
   - You can copy or move your configuration file to either location as needed.

### Run
#### Two-step UI (project → server)
```
./sshmenu
```
#### Flat filtered list
```
./sshmenu <search>
# Example:
./sshmenu App
```

## Config File: sshmenu.yaml
Example:
```yaml
global_command: pamssh {server}
projects:
  - name: Customer A
    servers:
      - name: server1
        description: App A Server 1
      - name: server2
        description: App B Server 2
  - name: Customer B
    servers:
      - name: server3
        description: App C Server 3
        command: ssh server3 -p 2222
```

## Navigation
- Use ↑/↓ to navigate
- Enter to select
- ⏻ Quit to exit
- ⬅ Back returns to project selection

### Build yourself
```
go build -o sshmenu sshmenu.go
```

## License
MIT
