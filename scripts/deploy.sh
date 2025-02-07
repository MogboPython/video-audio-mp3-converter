#!/bin/bash

# Usage: ./deploy.sh [dev|prod]

ENV=$1

if [ "$ENV" != "dev" ] && [ "$ENV" != "prod" ]; then
    echo "Please specify environment: dev or prod"
    exit 1
fi

# Set environment
export GO_ENV=$ENV

# Build the application
go build -o app cmd/server/main.go

# Start the application
./app 