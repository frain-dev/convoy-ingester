Convoy Ingester
=========

Convoy Ingester is a serverless function that acts as a middleware between API providers and Convoy. This enables consumers benefit from Convoy without providers taking advantage of Convoy yet. 

The plan is to make this first class support in Convoy, this repo is a prototype for that future.

How To
========
How to run the ingester locally and publish an event to it.

### Start Server
```bash
$ cd example

$ ./run_ingester.sh \
  <group-id> \
  <api-key> \
  <paystack-secret> \
  <convoy-paystack-app-id> \
```

### Push Request
```bash
curl --request POST \
--url "http://localhost:8080" \
--header "Content-Type: application/json" \
--header "X-Paystack-Signature: <signature>" \
--header "X-Forwarded-For: <insert-ip-address> \
--data '{"event": "charge.created"}'
```
You can generate dummy hash (here)[https://go.dev/play/p/NfFgzhtj-N]

