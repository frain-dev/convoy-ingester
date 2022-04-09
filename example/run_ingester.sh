#!/bin/sh

# Run Ingester
# This script assumes you're in the examples directory.

# Secret: Subomi
# App ID: 39e11a08-a15b-4453-8a69-273f43e39338

export CONVOY_URL=https://cloud.getconvoy.io/api/v1
export CONVOY_GROUP_ID=$1
export CONVOY_API_KEY=$2
export PAYSTACK_SECRET=$3
export CONVOY_PAYSTACK_APP_ID=$4

# Build Ingester
go build -o convoy-ingester ../cmd/main.go

# Run Ingester
./convoy-ingester
