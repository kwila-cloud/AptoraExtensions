# Aptora Extensions

Web server extensions that enhance the Aptora MS SQL Server database with additional functionality and improved user interfaces.

## Overview

Aptora Extensions provides a modern web interface for accessing and analyzing data from your Aptora database. Built for internal use, it focuses on optimizing tedious tasks and providing intuitive data visualization.

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
You can access at `localhost:80`

## Tech Stack

- **Frontend**: React 19.2 + TypeScript + Vite + Tailwind CSS
- **Backend**: Go 1.25 + Chi router
- **Database**: Microsoft SQL Server (Aptora + Extensions databases)
- **Deployment**: Single Go binary with embedded frontend

## Key Features

- Single executable deployment (no separate static files)
- Hot-reload development with instant feedback
- Secure read-only access to Aptora database
- Clean, minimal interface focused on usability
- Powerful data tables with sorting and filtering

## Architecture

The project uses a monorepo structure with:
- `/backend` - Go server with embedded frontend
- `/frontend` - React application
- Single `justfile` coordinates builds
- Environment-based configuration

## Development

Frontend runs on port 5173, backend proxies API calls and serves the React app in development. In production, the frontend is embedded into the Go binary for simplified deployment.

## Current Status

**Proof-of-Concept Phase**: Validating core database connectivity and basic web interface functionality.
