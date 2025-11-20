# Proof-of-concept

The very basic web server to prove that we can successfully pull data from the Aptora DB and display it in the user's web browser

## Task List

### Backend Creation

- [ ] Set up Go project following best practices
- [ ] Set up backend jobs in CI GitHub workflow
- [ ] Set up configuration system that makes it easy to configure the following from a single file
  - [ ] Aptora database URL
  - [ ] Aptora database port
  - [ ] Aptora database username
  - [ ] Aptora database password
  - [ ] Aptora Extensions database URL
  - [ ] Aptora Extensions database port
  - [ ] Aptora Extensions database username
  - [ ] Aptora Extensions database password
- [ ] Add backend getting started info to CONTRIBUTING.md

### Frontend Creation

- [ ] Set up React project following best practices
- [ ] Set up frontend jobs in CI GitHub workflow
- [ ] Set up backend to serve frontend using Go's `embed` package
  - [ ] Production: embed React build into Go binary
  - [ ] Development: `--dev-mode` flag to serve from disk with hot-reload
- [ ] Add frontend getting started info to CONTRIBUTING.md

### Build Set Up

- [ ] Create build script with a justfile
  - [ ] Build frontend (React production build)
  - [ ] Build backend (embed frontend assets into Go binary)
  - [ ] Output: single executable binary for deployment

### Deployment Script

- [ ] Create systemd service file for running the backend
- [ ] Create deployment script
  - [ ] Stop systemd service on the remote server (if it exists)
  - [ ] scp single binary executable to remote server
  - [ ] scp config file to remote server
  - [ ] Create and initialize systemd service on remote server (if it doesn't exist)
  - [ ] Start systemd service

### Backend Database Connection

- [ ] Create read-only connection to Aptora database
- [ ] Create read-write connection to Aptora Extensions database
- [ ] Create simple endpoint that allows querying employees from the DB
  - [ ] Only return ID and name
- [ ] Create simple endpoint that allows querying invoices from the DB
  - [ ] Parameters:
    - [ ] Date range (required)
    - [ ] Employee ID (optional)
  - [ ] Do not allow querying more than 500 invoices in one request

### Simple Frontend

- [ ] Create a single React page that fetches the invoices from last month
- [ ] Make it easy to filter by employee
