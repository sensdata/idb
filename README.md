# iDB

Self-hosted infrastructure management platform for servers, containers, files, services, firewall rules, and common operations.

[English](README.md) | [简体中文](README.zh-CN.md)

[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/sensdb/idb)](https://hub.docker.com/r/sensdb/idb)

iDB is built for developers and small teams who want one place to manage Linux hosts without introducing a large ops stack. It includes a web console, a local `center` service, and an `agent` used for host-side operations.

## Features

- Multi-host management with a single web console
- Host monitoring for CPU, memory, disk, network, and running processes
- Web terminal, file management, and service management
- Docker and Docker Compose operations
- Firewall management based on `nftables`
- SSH settings and certificate management
- Scheduled tasks, log viewing, and file sync workflows
- Built-in application deployment flows for common services

## Architecture

- `center`: control plane, API server, web UI host, task/log coordination
- `agent`: host-side runtime used for terminal, file, service, and system operations
- `frontend`: Vue-based web console
- `plugins`: optional app/plugin integrations

## Installation

### Quick Install

Install from GitHub directly:

```bash
sudo curl -fsSL https://raw.githubusercontent.com/sensdata/idb/main/scripts/install.sh | sudo bash
```

Install with the iDB mirror/proxy fallback explicitly enabled:

```bash
curl -fsSL https://dl.idb.net/github-raw/sensdata/idb/main/scripts/install.sh | sudo IDB_GITHUB_PROXY=https://dl.idb.net bash
```

If GitHub access is unstable, you can also download first and run locally:

```bash
curl -fsSL https://dl.idb.net/github-raw/sensdata/idb/main/scripts/install.sh -o install.sh
sudo IDB_GITHUB_PROXY=https://dl.idb.net bash install.sh
```

Notes:

- The install script already auto-detects GitHub connectivity and may fall back to `https://dl.idb.net`.
- The script requires `root` or `sudo`.
- After installation, the script prints the initial `admin` password once.

### Docker Install

Use GitHub release assets directly:

```bash
VERSION=$(curl -s https://api.github.com/repos/sensdata/idb/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
mkdir -p /var/lib/idb && cd /var/lib/idb
curl -fsSL "https://github.com/sensdata/idb/releases/download/${VERSION}/idb.env" -o .env
curl -fsSL "https://github.com/sensdata/idb/releases/download/${VERSION}/docker-compose.yaml" -o docker-compose.yaml
docker compose up -d
```

Use the iDB mirror/proxy for release assets:

```bash
VERSION=$(curl -s https://dl.idb.net/github-api/repos/sensdata/idb/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
mkdir -p /var/lib/idb && cd /var/lib/idb
curl -fsSL "https://dl.idb.net/github-releases/sensdata/idb/releases/download/${VERSION}/idb.env" -o .env
curl -fsSL "https://dl.idb.net/github-releases/sensdata/idb/releases/download/${VERSION}/docker-compose.yaml" -o docker-compose.yaml
docker compose up -d
```

### Build From Source

```bash
git clone --recurse-submodules https://github.com/sensdata/idb.git
cd idb
make deploy
```

`make deploy` builds the current checked-out source tree and installs both `idb` and `idb-agent` on the local machine.

What it does:

- builds the frontend and backend artifacts
- prepares keys, JWT material, and certificates if needed
- installs or updates the `center` service
- installs or updates the `agent` service

Important:

- `make deploy` is not restricted to `main`; it deploys the current branch and commit in your workspace
- it will overwrite the locally installed `idb` and `idb-agent` services
- for nontrivial upgrades, back up `/var/lib/idb/data` and `/etc/idb` first

The initial admin password is also written to:

```bash
/etc/idb/idb.env
```

You can retrieve it with:

```bash
sudo grep '^PASSWORD=' /etc/idb/idb.env
```

## Upgrade

Upgrade is handled by the release upgrade script. Direct GitHub:

```bash
curl -fsSL https://raw.githubusercontent.com/sensdata/idb/main/scripts/upgrade.sh -o upgrade.sh
sudo bash upgrade.sh
```

Upgrade with the iDB mirror/proxy explicitly enabled:

```bash
curl -fsSL https://dl.idb.net/github-raw/sensdata/idb/main/scripts/upgrade.sh -o upgrade.sh
sudo IDB_GITHUB_PROXY=https://dl.idb.net bash upgrade.sh
```

Current behavior:

- `center` upgrades first
- after `center` restarts, the default host agent is checked automatically
- if the local agent version is behind, it is upgraded automatically without requiring the user to trigger a manual host update

## Quick Start

1. Open the web console in your browser.
   Use `http://your-server-ip:9918` in a default deployment.
2. Log in with:
   Username: `admin`
   Password: the password printed by the installer or stored in `/etc/idb/idb.env`
3. Add remote hosts from the host management page.
4. Install or upgrade agents where needed.
5. Start using file management, terminal, service management, Docker operations, and app deployment.

## Repository Layout

```text
idb/
├── agent/       # Agent runtime
├── center/      # Center service and API
├── core/        # Shared constants, models, utilities
├── frontend/    # Web frontend
├── plugins/     # Plugin submodules and app integrations
├── scripts/     # Install / upgrade / uninstall scripts
└── README.md
```

## Development

### Requirements

- Go 1.25+
- Node.js 24+
- npm

### Backend

```bash
cd center
go run main.go start
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

### Source Deploy Testing

To validate a branch on a test machine:

```bash
git fetch origin
git switch your-branch
git pull
make deploy
```

Recommended post-deploy checks:

```bash
sudo systemctl status idb --no-pager -l
sudo systemctl status idb-agent --no-pager -l
sudo journalctl -u idb -n 100 --no-pager
sudo tail -n 100 /var/log/idb-agent/idb-agent.log
```

To confirm what was deployed:

```bash
git branch --show-current
git rev-parse --short HEAD
```

## Contributing

Contributions are welcome.

Basic workflow:

1. Fork the repository
2. Create a feature branch
3. Make focused changes
4. Add or update tests where appropriate
5. Submit a pull request

Guidelines:

- follow standard Go formatting and existing project conventions
- keep commits readable and scoped
- update documentation when behavior changes

## Documentation

- Website: https://idb.net
- Docs: https://idb.net/docs
- API docs: `http://your-server-ip:9918/api/v1/swagger/index.html`
- Plugin repository: https://github.com/sensdata/idb-plugins

To regenerate Swagger docs:

```bash
cd center
go generate .
```

## License

Apache License 2.0. See [LICENSE](LICENSE).
