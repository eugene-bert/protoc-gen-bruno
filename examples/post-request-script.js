// Example post-request script
// Log response and save data for next requests

console.log("Response status:", res.status);
console.log("Response body:", res.body);

// Save response data to collection variables
if (res && res.body) {
  if (res.body.id) {
    bru.setVar("last_created_id", res.body.id);
  }
  if (res.body.token) {
    bru.setVar("auth_token", res.body.token);
  }
}
