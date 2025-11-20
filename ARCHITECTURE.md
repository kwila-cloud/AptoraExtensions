# Architecture

## Technical Stack (High Level)

- **Frontend**: React v19.2 
  - React Router or Remix (need to determine best fit)
  - Some well-maintained package ()
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

- The interface should be intuitive and bare bones.
- Practicality trumps aesthetic in every design decision.

### Secure

- We need to be very sensitive about accessing the Aptora data.
- We do NOT want to allow unauthorized data access.
- We do NOT want to corrupt the Aptora database.

## To Do

- [ ] Determine if we should refine or add any more high level principles
- [ ] Figure out a good plan for combining React and Go into a single server with a single open port.
- [ ] Determine if react router or remix is a better fit.
- [ ] Determine good file structure to have backend, frontend, and deployments scripts in a single repo.
- [ ] Flesh out the proof-of-concept spec.
