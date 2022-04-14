#!/bin/sh

# Run Ingester
# This script assumes you're in the examples directory.

export GOOGLE_APPLICATION_CREDENTIALS=$1
export WEBHOOK_TOPIC=$2
export GOOGLE_CLOUD_PROJECT=$3
export PAYSTACK_SECRET=$4

# Build Ingester
go build -o convoy-ingester ../cmd/ingester/main.go

# Run Ingester
./convoy-ingester
