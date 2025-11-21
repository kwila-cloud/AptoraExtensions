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
- [x] Set up backend CI jobs in GitHub workflow (GitHub-hosted `ubuntu-latest` runners)
  - [x] `go vet` (Go static analysis)
  - [x] `go fmt` check (format verification)
  - [x] Unit tests with `go test`
- [x] Set up environment variable configuration (use real Aptora dev server per `docs/dev-server.md`)
  - [x] Use `godotenv` package for loading `.env` in dev mode
  - [x] Read environment variables with `os.Getenv()`
  - [x] Required variables (both databases on same SQL Server instance):
    - [x] `DB_HOST`, `DB_PORT` (shared by both databases)
    - [x] `APTORA_DB_NAME`, `APTORA_DB_USER`, `APTORA_DB_PASSWORD` (read-only)
    - [x] `EXTENSIONS_DB_NAME`, `EXTENSIONS_DB_USER`, `EXTENSIONS_DB_PASSWORD` (read-write)
- [x] Create `.env.example` file with all required variables
- [x] Add `.env` to `.gitignore`
- [x] Add backend getting started info to CONTRIBUTING.md

### Frontend Creation

- [x] Set up React v19.2 project with TypeScript and Vite
  - [x] Use Vite for build tooling (faster dev server, as mentioned in spec)
  - [x] Configure TypeScript for type safety
  - [x] No testing framework for POC (focus on functionality first)
- [x] Add React Router v7 for client-side routing
- [x] Set up frontend CI jobs in GitHub workflow
  - [x] ESLint (linting)
  - [x] Prettier check (format verification)
- [x] Set up backend to serve frontend using Go's `embed` package
  - [x] Production: embed React build into Go binary, listen on port 80
  - [x] Development: `--dev-mode` flag proxies frontend requests to Vite dev server (Go on :8080, Vite on :5173)
- [x] Add frontend getting started info to CONTRIBUTING.md

### Build Set Up

- [x] Create build script with a justfile
  - [x] Build frontend (React production build)
  - [x] Build backend (embed frontend assets into Go binary)
  - [x] Output: single executable binary for deployment

### Backend Database Connection

- [x] Add `microsoft/go-mssqldb` dependency to `go.mod`
- [x] Create database connection manager in `internal/database/`
  - [x] Connection pooling (max 10 connections per database)
  - [x] Read-only connection to Aptora database (Employees, Invoices tables)
  - [x] Read-write connection to Extensions database
  - [x] Retry logic: attempt connection every 30 seconds if initial connection fails
  - [x] Server starts even if databases unavailable (reports unhealthy, retries in background)
- [x] Create Extensions DB schema initialization
  - [x] `health_check` table: `id INT IDENTITY(1,1) PRIMARY KEY`, `timestamp DATETIME2`
  - [x] Insert new test row on each successful connection (always insert, never update)
- [ ] Update health check HTTP endpoint (`GET /health`)
  - [ ] Returns `{"status": "healthy"}` (200 OK) if both databases connected
  - [ ] Returns `{"status": "unhealthy", "error": "..."}` (503) if either database unavailable
- [ ] Create employees endpoint (`GET /api/employees`)
  - [ ] Query `Employees` table: `id` and `Name` columns
  - [ ] Response format: `{"employees": [{"id": 1, "name": "John Doe"}, ...]}`
  - [ ] Error format: `{"error": "error message"}`
- [ ] Create invoices endpoint (`GET /api/invoices`)
  - [ ] Required query params: `start_date`, `end_date` (YYYY-MM-DD format, inclusive range)
  - [ ] Optional query param: `employee_id` (integer)
  - [ ] Query `Invoices` table: `id`, `Date`, `RepID`, `Total` columns
  - [ ] Filter: `Date >= start_date AND Date <= end_date`, optionally filter by `RepID = employee_id`
  - [ ] Response format: `{"invoices": [{"id": 123, "date": "2025-01-15", "rep_id": 5, "total": 1500.00}, ...]}`
  - [ ] Error format: `{"error": "error message"}`
  - [ ] Validation: Return error if query would return >500 invoices  (error message should mention using a narrower filter)

### Simple Frontend

- [ ] Set up Tailwind CSS in React project
- [ ] Remove current unnecessary CSS
- [ ] Set up TanStack Table
- [ ] Create a single React page that fetches invoices from last month
  - [ ] Display all invoice fields returned from DB (refine later)
  - [ ] Use TanStack Table for spreadsheet-like display
  - [ ] Implement sorting functionality
  - [ ] Implement filtering by employee
  - [ ] Keep UI minimal but polished (align with Low Friction principle)

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

