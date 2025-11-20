# Architecture

## Technical Stack (High Level)

- **Frontend**: React v19.2 
  - React Router v7 for client-side routing
  - Vite for build tooling and dev server
  - Tailwind CSS for styling
  - TanStack Table for data grid functionality
- **Backend**: Go v1.25
  - Chi v5.2 for routing and middleware
  - microsoft/go-mssqldb v1.9 for SQL driver
- **Database**: Microsoft SQL Server
  - Main database - Aptora (read-only connection)
  - Auxilary database - used for persisting state used by the extensions
- **Deployment**: Custom script to copy files over SSH and restart systemd service

## High Level Principles

### Simplicity

- The server and frontend should be as simple as possible.
- The goal is to make a tool that is hyper-focused on optimizing tedious tasks in Aptora.

### Low Friction

- The interface should be intuitive and minimal.
- Minimize clicks and cognitive load.
- Use clean, functional UI without unnecessary decoration.
- Practicality trumps aesthetic - polished is good, but ugliness that hinders usability should be avoided.

### Maintainability

- Code should be easy to understand, debug, and extend.
- Optimize for minimal engineering time while maintaining code quality.
- Clear structure and documentation to support long-term evolution.

### Secure

- We need to be very sensitive about accessing the Aptora data.
- We do NOT want to allow unauthorized data access.
- We do NOT want to corrupt the Aptora database.

## Configuration

### Environment Variables
- Configuration via environment variables (12-factor app style)
- Production: systemd service uses `EnvironmentFile=/opt/aptora-extensions/.env`
- Development: `godotenv` package loads `.env` file in dev mode
- No config parsing library needed - use Go's built-in `os.Getenv()`
- Required variables:
  - `DB_HOST`, `DB_PORT` (shared - both databases on same SQL Server instance)
  - `APTORA_DB_NAME`, `APTORA_DB_USER`, `APTORA_DB_PASSWORD` (read-only connection)
  - `EXTENSIONS_DB_NAME`, `EXTENSIONS_DB_USER`, `EXTENSIONS_DB_PASSWORD` (read-write connection)

### Benefits
- Simplest approach - no YAML/TOML/JSON parsing
- Systemd handles parsing the `.env` file
- Easy deployment: just scp the `.env` file
- Standard across all platforms

## Project Structure

```
/
├── backend/           # Go server code
│   ├── cmd/          # Application entry points
│   ├── internal/     # Private application code
│   └── config/       # Configuration handling
├── frontend/         # React application
│   ├── src/          # React source code
│   ├── public/       # Static assets
│   └── dist/         # Build output (embedded into Go binary)
├── deploy/           # Deployment scripts and systemd configs
├── justfile          # Build automation (frontend + backend)
└── .github/
    └── workflows/    # CI/CD pipelines
```

### Rationale
- **Clear separation**: Frontend and backend are distinct directories
- **Simple navigation**: Easy to find what you need
- **Deployment isolation**: `/deploy` keeps deployment concerns separate
- **Single justfile**: Coordinates builds across frontend and backend

## Frontend Serving Strategy

### Production
- Go's `embed` package bundles React build into the binary
- Single executable deployment - no separate static files needed
- Frontend and backend versions always in sync
- Server listens on hardcoded port 80 (HTTP only - internal network use only)

### Development
- `--dev-mode` flag makes Go proxy frontend requests to Vite dev server
- Go backend runs on port 8080, proxies `/` requests to Vite on port 5173
- Enables hot-reload for fast iteration with Vite's instant feedback
- API requests handled directly by Go backend

### Benefits
- Simplifies deployment (single binary to copy)
- Eliminates version mismatch issues
- Maintains fast dev workflow with hot-reload
- No need for separate nginx/reverse proxy in production

