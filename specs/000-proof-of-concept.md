# Proof-of-concept

**Status**: in-progress

## Description

The very basic bare-bones web server to prove that we can successfully pull data from the Aptora MS SQL Server database and display it in the user's web browser. This POC validates the core technical stack (Go backend + React frontend + MSSQL database) and deployment workflow before building more complex features.

## Design Decisions

### Configuration System
- **Decision**: Environment variables with `.env` file
- **Rationale**: Simplest approach - no config parsing library needed, systemd handles `.env` parsing via `EnvironmentFile`
- **Development**: Use `godotenv` package to load `.env` in dev mode
- **Deployment**: scp `.env.production` to server as `.env`

### Project Structure
- **Decision**: Monorepo with `/backend`, `/frontend`, and `/deploy` directories at root
- **Rationale**: Clear separation of concerns, easy navigation, simple for small team
- **Build coordination**: Single justfile at root orchestrates frontend and backend builds
- **Deployment**: Deployment scripts and systemd configs live in `/deploy` directory

### Frontend Serving Strategy
- **Decision**: Use Go's `embed` package to bundle React build into the binary
- **Pros**: Single executable deployment, version consistency, simpler systemd service
- **Cons**: Larger binary size, full recompile for frontend changes
- **Mitigation**: `--dev-mode` flag proxies frontend requests to Vite dev server (Go on :8080, Vite on :5173)
- **Production**: Server listens on hardcoded port 80 (HTTP only - internal network use only)

### Frontend Routing
- **Decision**: React Router v7 for client-side routing
- **Rationale**: Sufficient for multiple pages (invoices, employees, reports), simpler than Remix, no SSR needed for internal tool
- **Alternative considered**: Remix (rejected - overkill without SEO requirements, adds complexity with Go already handling backend)

### Frontend Styling & Components
- **Decision**: Tailwind CSS for styling, TanStack Table for data grid
- **Rationale**: 
  - Tailwind aligns with principles (minimal, maintainable, fast iteration)
  - No component library for POC - add as needed to avoid bloat and duplicate code
  - TanStack Table is free/open-source with no paid tiers, headless (works with Tailwind), powerful sorting/filtering
- **Alternative considered**: AG Grid (rejected - has paid tiers, vendor lock-in concerns)

### Error Handling
- **Decision**: Simple error handling for POC - log errors and return 500s
- **Rationale**: Focus on happy path first, can add structured errors later
- **Future**: Add proper HTTP status codes and error responses post-POC

### Logging
- **Decision**: Use Go's `slog` package (structured logging)
- **Rationale**: Standard library (no dependencies), structured output, sufficient for needs

### Authentication
- **Decision**: No authentication for POC
- **Rationale**: Internal network only, focus on core functionality first
- **Future**: See spec 001-implement-basic-auth.md for post-POC auth plan

## Task List

### Backend Creation

- [x] Set up Go project following best practices
  - [x] Structure modules with `cmd/server` entrypoint and `internal/...` packages
  - [x] Use `slog` for structured logging
- [ ] Set up backend CI jobs in GitHub workflow (GitHub-hosted `ubuntu-latest` runners)
  - [ ] `go vet` (Go static analysis)
  - [ ] `go fmt` check (format verification)
  - [ ] Unit tests with `go test`
- [ ] Set up environment variable configuration (use real Aptora dev server per `docs/dev-server.md`)
  - [ ] Use `godotenv` package for loading `.env` in dev mode
  - [ ] Read environment variables with `os.Getenv()`
  - [ ] Required variables (both databases on same SQL Server instance):
    - [x] `DB_HOST`, `DB_PORT` (shared by both databases)
    - [x] `APTORA_DB_NAME`, `APTORA_DB_USER`, `APTORA_DB_PASSWORD` (read-only)
    - [x] `EXTENSIONS_DB_NAME`, `EXTENSIONS_DB_USER`, `EXTENSIONS_DB_PASSWORD` (read-write)
- [ ] Create `.env.example` file with all required variables
- [ ] Add `.env` to `.gitignore`
- [ ] Add backend getting started info to CONTRIBUTING.md

### Frontend Creation

- [ ] Set up React project following best practices
- [ ] Set up frontend CI jobs in GitHub workflow
  - [ ] ESLint (linting)
  - [ ] Prettier check (format verification)
- [ ] Set up backend to serve frontend using Go's `embed` package
  - [ ] Production: embed React build into Go binary, listen on port 80
  - [ ] Development: `--dev-mode` flag proxies frontend requests to Vite dev server (Go on :8080, Vite on :5173)
- [ ] Add frontend getting started info to CONTRIBUTING.md

### Build Set Up

- [ ] Create build script with a justfile
  - [ ] Build frontend (React production build)
  - [ ] Build backend (embed frontend assets into Go binary)
  - [ ] Output: single executable binary for deployment

### Deployment Script

- [ ] Create systemd service file for running the backend
  - [ ] Use `EnvironmentFile=/opt/aptora-extensions/.env` to load config
  - [ ] Set `WorkingDirectory=/opt/aptora-extensions`
  - [ ] Configure restart policy and logging
- [ ] Create deployment script
  - [ ] Stop systemd service on the remote server (if it exists)
  - [ ] scp single binary executable to remote server
  - [ ] scp `.env.production` to remote server as `.env`
  - [ ] Create and initialize systemd service on remote server (if it doesn't exist)
  - [ ] Start systemd service

### Backend Database Connection

- [ ] Create read-only connection to Aptora database
- [ ] Create read-write connection to Aptora Extensions database
- [ ] Create simple health check table in Extensions DB to verify write access
  - [ ] Table: `health_check` with `id` and `timestamp` columns
  - [ ] Insert test row on startup to verify connection works
- [ ] Create health check HTTP endpoint
  - [ ] `GET /health` returns 200 OK if both database connections are healthy
  - [ ] Returns 503 Service Unavailable if either database is unreachable
  - [ ] Include basic status info in response (e.g., database connectivity status)
- [ ] Create simple endpoint that allows querying employees from the DB
  - [ ] Only return ID and name
- [ ] Create simple endpoint that allows querying invoices from the DB
  - [ ] Parameters:
    - [ ] Date range (required)
    - [ ] Employee ID (optional)
  - [ ] Return all invoice fields from DB (can refine later based on frontend needs)
  - [ ] Do not allow querying more than 500 invoices in one request

### Simple Frontend

- [ ] Set up Tailwind CSS in React project
- [ ] Set up TanStack Table
- [ ] Create a single React page that fetches invoices from last month
  - [ ] Display all invoice fields returned from DB (refine later)
  - [ ] Use TanStack Table for spreadsheet-like display
  - [ ] Implement sorting functionality
  - [ ] Implement filtering by employee
  - [ ] Keep UI minimal but polished (align with Low Friction principle)
