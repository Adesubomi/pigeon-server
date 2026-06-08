# Pigeon Server

## GitHub OAuth

Create a GitHub OAuth App at `Settings > Developer settings > OAuth Apps`.

For local development, configure:

```text
Homepage URL: http://localhost:3000
Authorization callback URL: http://localhost:3000/auth/callback
```

Then copy the app's client ID and generated client secret into `.env`:

```dotenv
GITHUB_CLIENT_ID=your-client-id
GITHUB_CLIENT_SECRET=your-client-secret
WEB_APP_URL=http://localhost:3000
```

Restart the server after changing `.env`. `GET /auth/github` returns the GitHub
authorization URL. A `503 auth.github_not_configured` response means one or both
credentials are still missing from the server process.

GitHub redirects to the web client's `/auth/callback` page. The web client
validates OAuth state and sends the authorization code to
`POST /auth/github/exchange`; the stateless API does not expose its own callback
route.

`WEB_APP_URL/auth/callback` is trusted automatically. Set `OAUTH_REDIRECT_ALLOWLIST` to a comma-separated list of additional exact callback URLs when multiple web deployments are required. Authorization and code exchange requests reject any other `redirect_uri`.

CORS allows the exact origin from `WEB_APP_URL` and supports `Authorization` and `Content-Type` for the SPA bearer-token flow.

`POST /auth/logout` intentionally returns `204 No Content` without revoking a token. Access tokens are self-contained JWTs and remain technically valid until expiry; the web client always removes its local token on logout.
