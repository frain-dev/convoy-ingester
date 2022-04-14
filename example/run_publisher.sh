#!/bin/sh

# Run Publisher
# This script assumes you're in the examples directory.

export GOOGLE_APPLICATION_CREDENTIALS=$1
export WEBHOOK_TOPIC=$2
export GOOGLE_CLOUD_PROJECT=$3
export CONVOY_GROUP_ID=$4
export CONVOY_API_KEY=$5
export CONVOY_PAYSTACK_APP_ID=$6
export CONVOY_URL=https://cloud.getconvoy.io/api/v1

# Build publisher
go build -o convoy-publisher ../cmd/publisher/main.go

# Run publisher
./convoy-publisher
