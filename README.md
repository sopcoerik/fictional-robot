# DEManager - Development Environment Manager

A lightweight terminal UI tool for orchestrating and monitoring multiple local services during development. Start, stop, and manage your entire dev stack from a single dashboard.

## Why DEManager?

Managing multiple services (Redis, backend API, frontend) across different terminals is tedious and error-prone. DEManager solves this by:

- **Single executable**: No Docker, no containers, just native processes
- **Dependency handling**: Services start in the correct order automatically
- **Real-time logs**: View each service's output in separate tabs
- **Control**: Start, stop, and restart services individually or all at once
- **Health checks**: Automatic TCP port checks to verify services are running

## Installation

### Prerequisites
- Go 1.24 or later

### Build from Source

```bash
git clone https://github.com/sopcoerik/fictional-robot.git
cd fictional-robot
go build -o demanager ./cmd/demanager
```

This creates a `demanager` executable in your current directory.

## Quick Start

### 1. Create a Configuration File

Create a `devenv.yaml` in your project:

```yaml
services:
  redis:
    command: "redis-server"
    port: 6379
  
  backend:
    command: "cd backend && npm run dev"
    port: 3000
    depends_on: [redis]
  
  frontend:
    command: "cd frontend && npm run dev"
    port: 5173
    depends_on: [backend]
```

### 2. Run DEManager

```bash
./demanager
```

The dashboard will launch automatically in your terminal.

## Usage

### Navigation

- **Arrow Keys (Left/Right)**: Switch between service tabs
- **Tab / Shift+Tab**: Navigate between buttons
- **Click on Tab**: Jump directly to a service
- **Enter / Space**: Activate focused button
- **q / Ctrl+C**: Quit

### Controls

| Button | Action |
|--------|--------|
| **Stop** | Stop the currently selected service |
| **Start** | Start the currently selected service |
| **Restart All** | Stop all services and restart them in dependency order |

### Example Workflow

1. Dashboard opens with all services starting in order (Redis → Backend → Frontend)
2. View each service's logs in its tab
3. If backend crashes, click the Backend tab, hit **Start** to restart it
4. Hit **Restart All** to reset everything
5. Press `q` to cleanly shutdown all services

## Architecture

### Core Components

- **parser.go**: Loads and validates `devenv.yaml`
- **sorter.go**: Resolves service dependencies using topological sort
- **starter.go**: Spawns and monitors processes, captures logs
- **main.go**: Orchestrates app state, manages contexts
- **ui.go**: BubbleTea-based terminal dashboard

### Execution Flow

1. Parse config and validate dependencies
2. Start services in dependency order
3. Monitor health (TCP port checks)
4. Collect stdout/stderr logs per service
5. Render interactive TUI dashboard
6. Handle user input (stop/start/restart)
7. Graceful shutdown on exit (Ctrl+C)

### Context Hierarchy

```
ParentContext (OS signals: Ctrl+C, SIGTERM)
  └── GlobalContext (current service cycle)
      ├── ServiceContext (Redis)
      ├── ServiceContext (Backend)
      └── ServiceContext (Frontend)
```

Each service context can be cancelled independently, while the global context controls all services together (used by Restart All).

## Features

✅ YAML configuration parsing  
✅ Automatic dependency resolution (detects circular dependencies)  
✅ TCP health checks on startup  
✅ Real-time log capture (stdout/stderr)  
✅ Per-service log viewing (last 200 entries)  
✅ Individual service control (start/stop)  
✅ Batch restart (restart all services)  
✅ Graceful shutdown (SIGINT handling)  
✅ Cross-platform (Linux verified; Windows/macOS support pending)

## Configuration Reference

### devenv.yaml Format

```yaml
services:
  service_name:
    command: "command to run"      # Required: shell command
    port: 1234                      # Required: port for health check
    depends_on: [other_service]    # Optional: list of dependencies
```

### Rules

- Service names must be unique
- Circular dependencies are detected and rejected
- Services in `depends_on` must exist in the config
- Health checks fail if the port isn't reachable after 5 seconds

## Limitations

- Log view shows last 30 lines (200 stored in memory)
- No log scrolling yet (planned for v1.1)
- No SQLite persistence (planned for v1.1)
- No CPU/RAM metrics (planned for v1.1)
- Cross-platform signals may need tuning on Windows

## Troubleshooting

### "Health check timeout"
The service started but isn't listening on its port within 5 seconds. Check:
- Is the service actually running? (Look at logs)
- Is the port number correct in config?
- Is the service binding to localhost?

### "Circular dependency detected"
Your `depends_on` creates a cycle. Example:
```yaml
services:
  a:
    depends_on: [b]
  b:
    depends_on: [a]  # ← This creates a cycle
```

### UI doesn't render correctly
Try resizing your terminal or running in fullscreen. BubbleTea needs sufficient space.

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

```
.
├── main.go           # App orchestration
├── ui.go             # Terminal UI
├── parser.go         # Config parsing
├── sorter.go         # Dependency resolution
├── starter.go        # Process management
├── devenv.yaml       # Example config
└── go.mod
```

## Future Plans

- [ ] Log scrolling and persistence
- [ ] SQLite event history (start/stop/crash timestamps)
- [ ] Service status indicators (running/stopped/crashed)
- [ ] CPU and memory metrics
- [ ] Web dashboard alternative
- [ ] Configuration hot-reload
- [ ] Custom health check endpoints

## Contributing

Bug reports and feature requests welcome! Please open an issue on GitHub.

## License

MIT License - see LICENSE file for details.

---

**Built with:**
- Go (backend, process management)
- BubbleTea (terminal UI)
- Lipgloss (UI styling)
- YAML v3 (configuration)

**Author:** [Your Name]  
**Repository:** https://github.com/sopcoerik/fictional-robot