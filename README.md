# Aptora Extensions

Aptora Extensions provides a modern web interface for accessing and analyzing data from your [Aptora](http://aptora.com) database. Built to be self-hosted on an internal network, it focuses on optimizing tedious tasks and providing intuitive data visualization.

## Quick Start

### Prerequisites

Before running the application, you need to set up database credentials:

#### 1. Create Read-Only User for Aptora Database

Connect to your SQL Server instance and run:

```sql
-- Create read-only login
CREATE LOGIN aptora_readonly WITH PASSWORD = 'your_secure_password';

-- Switch to Aptora database
USE [YourAptoraDatabase];

-- Create user from login
CREATE USER aptora_readonly FOR LOGIN aptora_readonly;

-- Grant read-only access
ALTER ROLE db_datareader ADD MEMBER aptora_readonly;

-- Verify read-only access (optional)
-- This should fail with permission denied:
-- EXECUTE AS USER = 'aptora_readonly';
-- CREATE TABLE test (id INT);
-- REVERT;
```

#### 2. Create Read-Write User for Extensions Database

```sql
-- Create read-write login
CREATE LOGIN aptora_extensions WITH PASSWORD = 'your_secure_password';

-- Create Extensions database if it doesn't exist
IF NOT EXISTS (SELECT * FROM sys.databases WHERE name = 'AptoraExtensions')
BEGIN
    CREATE DATABASE AptoraExtensions;
END

-- Switch to Extensions database
USE AptoraExtensions;

-- Create user from login
CREATE USER aptora_extensions FOR LOGIN aptora_extensions;

-- Grant read-write access
ALTER ROLE db_datareader ADD MEMBER aptora_extensions;
ALTER ROLE db_datawriter ADD MEMBER aptora_extensions;
ALTER ROLE db_ddladmin ADD MEMBER aptora_extensions;
```

### Development

```bash
# Copy environment template and edit with your database credentials
cp .env.example .env

# Install frontend dependencies
cd frontend && npm install && cd ..

# Start development servers (Go backend + Vite frontend)
just dev
```
You can access at `localhost:8080`

### Production Build

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your database credentials
just build-backend

# Must use sudo to listen on port 80
sudo ./aptora-extensions
```
You can access at `0.0.0.0:80` (or from any other device on the network)

## Key Features

- Single executable deployment (no separate static files)
- Hot-reload development with instant feedback
- Secure read-only access to Aptora database
- Clean, minimal interface focused on usability
- Powerful data tables with sorting and filtering

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md).
