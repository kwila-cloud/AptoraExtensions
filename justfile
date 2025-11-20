#!/usr/bin/env just --justfile

# Build the frontend and copy to backend assets
build-frontend:
    cd frontend && npm ci && npm run build
    mkdir -p backend/internal/server/built-frontend
    cp -r frontend/dist/* backend/internal/server/built-frontend/

# Build the backend (embed frontend assets)
build-backend: build-frontend
    cd backend && go build -o ../aptora-extensions ./cmd/server

# Run development mode (frontend dev server + backend proxy)
dev:
    # Start frontend dev server in background
    cd frontend && npm run dev &
    # Start backend in dev mode
    cd backend && go run ./cmd/server --dev

# Clean build artifacts
clean:
    rm -rf frontend/dist
    rm -rf backend/internal/server/built-frontend
    rm -f aptora-extensions
