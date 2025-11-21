# Production Server

Information for setting up the production server.

## Create VM

- Create an Ubuntu LTS VM on your host (Hyper-V, ProxMox, etc.)
  - ISOs are available [here](https://ubuntu.com/download/server)
  - For ProxMox, you can use this [helper script](https://community-scripts.github.io/ProxmoxVE/scripts?id=ubuntu2404-vm)
    - See [here](https://github.com/community-scripts/ProxmoxVE/discussions/272) for useful configuration information.
    - You should not need to run `parted` to expand the disk - it seems to do it automatically
    - Do not follow the SSH config - the default config should allow pubkey access
    - You may need to run `sudo systemctl enable ssh`

## Remote access

- Set up SSH access
  - Set `PubkeyAuthentication yes` in `/etc/ssh/sshd_config`
- Set up VPN access
- Add username and identify key configuration for the host to `~/.ssh/config`
- Verify you can access outside of network using only `ssh <host-name>`

## Configuration

- Set up users on production DB using instructions from [README](../README.md)
- Set up `.env.production` with correct database host and credentials

## Deployment

Run `just deploy <host-name>`
