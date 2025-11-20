# Architecture

## Technical Stack (High Level)

- **Frontend**: React v19.2 
  - React Router or Remix (need to determine best fit)
- **Backend**: Go v1.25
  - Chi v5.2 for routing and middleware
  - microsoft/go-mssqldb v1.9 for SQL driver
- **Database**: Microsoft SQL Server (the Aptora database)
- **Deployment**: Custom script to copy files over SSH and restart systemd service

## To Do

- [ ] Figure out a good plan for combining React and Go into a single server with a single open port.
- [ ] Determine if react router or remix is a better fit.
- [ ] Determine good file structure to have backend, frontend, and deployments scripts in a single repo.
- [ ] Flesh out the proof-of-concept spec.
