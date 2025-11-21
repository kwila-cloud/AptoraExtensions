#!/bin/bash
set -euo pipefail

if [ $# -eq 0 ]; then
    echo "Error: Hostname is required"
    echo "Usage: $0 <hostname>"
    exit 1
fi

HOST="$1"
echo "Rolling back deployment on $HOST..."

# Check if backup exists
if ! ssh "$HOST" "[ -d /opt/aptora-extensions.backup ]"; then
    echo "✗ Error: No backup found at /opt/aptora-extensions.backup"
    exit 1
fi

# Stop service
echo "Stopping service..."
ssh "$HOST" "systemctl stop aptora-extensions"

# Restore from backup
echo "Restoring from backup..."
ssh "$HOST" "rm -rf /opt/aptora-extensions && mv /opt/aptora-extensions.backup /opt/aptora-extensions"

# Start service
echo "Starting service..."
ssh "$HOST" "systemctl start aptora-extensions"

# Wait and verify
echo "Waiting for service to start..."
sleep 5

echo "Verifying rollback..."
if ssh "$HOST" "systemctl is-active aptora-extensions && curl -f http://localhost/health"; then
    echo "✓ Rollback successful!"
    echo "View logs: ssh $HOST journalctl -u aptora-extensions -f"
else
    echo "✗ Rollback failed - service is not healthy"
    echo "Check logs: ssh $HOST journalctl -u aptora-extensions -n 50"
    exit 1
fi