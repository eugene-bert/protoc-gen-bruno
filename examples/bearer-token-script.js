// Fetch bearer token from auth server
const res = await req({
  url: bru.getEnvVar("auth_server_url") || "http://localhost:1234",
  method: "GET"
});

// Set the bearer token from response
if (res && res.body && res.body.token) {
  bru.setVar("bearer_token", res.body.token);
}
