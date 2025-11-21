# Production Server

Information for setting up the production server.

## Create VM

- Create an Ubuntu LTS VM on your host (Hyper-V, ProxMox, etc.)
  - ISOs are available here: https://ubuntu.com/download/server

## Remote access

- Set up SSH access
- Set up VPN access
- Add username and identify key configuration for the host to `~/.ssh/config`
- Verify you can access outside of network using only `ssh <host-name>`

## Configuration

- Set up users on production DB using instructions from [README](../README.md)
- Set up `.env.production` with correct database host and credentials

## Deployment

Run `just deploy <host-name>`
