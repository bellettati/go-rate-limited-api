The API exposes a single endpoint `POST /request`.

Clients are defined using the `X-API-Key` request header.

For each request, the service decides whether the client is allowed to proceed based on a fixed rate limit of 10 requests per minute.

If the request is allowed, the API responds with `200 OK`.

If the rate limit is exceeded, the API responds with `429 Too Many Requests`.
