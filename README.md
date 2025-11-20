# Aptora Extensions

Aptora Extensions provides a modern web interface for accessing and analyzing data from your [Aptora](http://aptora.com) database. Built to be self-hosted on an internal network, it focuses on optimizing tedious tasks and providing intuitive data visualization.

## Quick Start

### Development

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your database credentials
# Install frontend dependencies
cd frontend && npm install

# Start development servers (Go backend + Vite frontend)
just dev
```
You can access at `localhost:8080`

### Production Build

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your database credentials
just build-backend

# Must use sudo to listen on port 80
sudo ./aptora-extensions
```
You can access at `0.0.0.0:80` (or from any other device on the network)

## Key Features

- Single executable deployment (no separate static files)
- Hot-reload development with instant feedback
- Secure read-only access to Aptora database
- Clean, minimal interface focused on usability
- Powerful data tables with sorting and filtering

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md).
