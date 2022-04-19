Convoy Ingester
=========

Convoy Ingester is a serverless function that acts as a middleware between API providers and Convoy. This enables consumers benefit from Convoy without providers taking advantage of Convoy yet. 

The plan is to make this first class support in Convoy, this repo is a prototype for that future.

### Functions 

#### WebhookEndpoint
This function receives webhook events from the provider E.g. Paystack. acks the event and publishes the event to a pub/sub topic. To configure this function 
set the environment variable - `WEBHOOK_ENDPOINT_ENV_VARS` in GitHub actions with:

```bash
WEBHOOK_TOPIC=<insert-topic>,GOOGLE_CLOUD_PROJECT=<insert-project-id>,PAYSTACK_SECRET=<insert-paystack-secret>
```

#### PushToConvoy
This function is triggered from the pub/sub topic earlier and pushes to Convoy. To configure this function set environment variable - `PUSH_TO_CONVOY_ENV_VARS` in GitHub actions with:

```bash
WEBHOOK_TOPIC=<insert-topic>,GOOGLE_CLOUD_PROJECT=<insert-project-id>,CONVOY_GROUP_ID=<insert-group-id>,CONVOY_API_KEY=<insert-api-key>,CONVOY_PAYSTACK_APP_ID=<insert-app-id>
```

### How To
To run this function, you need to fork the repository. Follow this [article](https://www.honeybadger.io/blog/building-testing-and-deploying-google-cloud-functions-with-ruby/) to deploy these functions to Google Cloud Functions

### Push Request
```bash
curl --request POST \
--url "http://localhost:8080" \
--header "Content-Type: application/json" \
--header "X-Paystack-Signature: <signature>" \
--header "X-Forwarded-For: <insert-ip-address> \
--data '{"event": "charge.created"}'
```
You can generate dummy hash [here](https://go.dev/play/p/NfFgzhtj-N)
