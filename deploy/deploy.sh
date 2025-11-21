#!/bin/bash
set -euo pipefail

if [ $# -eq 0 ]; then
    echo "Error: Hostname is required"
    echo "Usage: $0 <hostname>"
    exit 1
fi

HOST="$1"
echo "Deploying to $HOST..."

# Create backup of existing binary (if exists)
echo "Creating backup of existing binary..."
ssh "$HOST" "mkdir -p /opt/aptora-extensions && [ -f /opt/aptora-extensions/aptora-extensions ] && cp /opt/aptora-extensions/aptora-extensions /opt/aptora-extensions/aptora-extensions.backup || true"

# Stop service (ignore if not running)
echo "Stopping service..."
ssh "$HOST" "systemctl stop aptora-extensions || true"

# Copy files
echo "Copying files to server..."
scp ./aptora-extensions "$HOST":/opt/aptora-extensions/
scp ./.env.production "$HOST":/opt/aptora-extensions/.env
scp ./deploy/aptora-extensions.service "$HOST":/etc/systemd/system/

# Set permissions
echo "Setting permissions..."
ssh "$HOST" "chown root:root /opt/aptora-extensions/aptora-extensions /opt/aptora-extensions/.env && chmod 755 /opt/aptora-extensions/aptora-extensions && chmod 600 /opt/aptora-extensions/.env"

# Enable and start service
echo "Enabling and starting service..."
ssh "$HOST" "systemctl daemon-reload && systemctl enable aptora-extensions && systemctl start aptora-extensions"

# Wait and verify
echo "Waiting for service to start..."
sleep 5

echo "Verifying deployment..."
if ssh "$HOST" "systemctl is-active aptora-extensions && curl -f http://localhost/health"; then
    echo "✓ Deployment successful!"
    echo "View logs: ssh $HOST journalctl -u aptora-extensions -f"
    echo "Rollback command (if needed): ssh $HOST 'systemctl stop aptora-extensions && [ -f /opt/aptora-extensions/aptora-extensions.backup ] && cp /opt/aptora-extensions/aptora-extensions.backup /opt/aptora-extensions/aptora-extensions && systemctl start aptora-extensions'"
else
    echo "✗ Deployment failed - service is not healthy"
    echo "Check logs: ssh $HOST journalctl -u aptora-extensions -n 50"
    echo "Rollback command: ssh $HOST 'systemctl stop aptora-extensions && [ -f /opt/aptora-extensions/aptora-extensions.backup ] && cp /opt/aptora-extensions/aptora-extensions.backup /opt/aptora-extensions/aptora-extensions && systemctl start aptora-extensions'"
    exit 1
fi