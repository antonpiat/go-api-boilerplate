# go-api-boilerplate

Domain-modular Gin REST API with Postgres, Redis, JWT/RBAC, Zap, Swagger, and Prometheus.

## Quick start

```bash
docker compose up --build
```

Then open:

| What | URL |
| --- | --- |
| API (redirects to docs) | http://localhost:8080 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| Prometheus UI | http://localhost:9090 |
| App metrics | http://localhost:8080/metrics |
| Liveness | http://localhost:8080/health/live |
| Readiness | http://localhost:8080/health/ready |

Auth and user routes are JSON APIs. Opening them in a browser sends `GET` and will 404 — use curl, httpie, or Swagger **Try it out**.

Without Docker:

```bash
docker compose up postgres redis -d
make migrate-up
make run
```

## How to use the API

All JSON bodies use `Content-Type: application/json`. Successful responses look like:

```json
{"success": true, "data": {}, "meta": {"page": 1, "per_page": 20, "total": 0}}
```

Errors:

```json
{"success": false, "error": {"code": "UNAUTHORIZED", "message": "..."}}
```

### 1. Register

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"password123"}'
```

`201` returns `access_token`, `refresh_token`, `token_type`, and `expires_in`.

### 2. Login

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"password123"}'
```

Save the access token:

```bash
TOKEN='<access_token>'
REFRESH='<refresh_token>'
```

### 3. Current user

```bash
curl -s http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"
```

Update email or password:

```bash
curl -s -X PATCH http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"email":"new@example.com"}'
```

### 4. Refresh and logout

Access tokens expire (default 15m). Rotate them:

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
  -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$REFRESH\"}"
```

Logout revokes the refresh token and blacklists the current access token:

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$REFRESH\"}"
```

### 5. Admin user CRUD

New accounts get role `user`. Promote one in Postgres, then log in again:

```sql
UPDATE users SET role = 'admin' WHERE email = 'you@example.com';
```

```bash
curl -s 'http://localhost:8080/api/v1/users?page=1&per_page=20' \
  -H "Authorization: Bearer $TOKEN"

curl -s http://localhost:8080/api/v1/users/<user-uuid> \
  -H "Authorization: Bearer $TOKEN"

curl -s -X PATCH http://localhost:8080/api/v1/users/<user-uuid> \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"role":"admin"}'

curl -s -X DELETE http://localhost:8080/api/v1/users/<user-uuid> \
  -H "Authorization: Bearer $TOKEN"
```

You can also drive the same calls from Swagger: Authorize with `Bearer <access_token>`.

## Endpoints

| Method | Path | Access |
| --- | --- | --- |
| POST | `/api/v1/auth/register` | public |
| POST | `/api/v1/auth/login` | public |
| POST | `/api/v1/auth/refresh` | public |
| POST | `/api/v1/auth/logout` | authenticated |
| GET/PATCH | `/api/v1/users/me` | authenticated |
| GET | `/api/v1/users` | admin |
| GET/PATCH/DELETE | `/api/v1/users/:id` | admin |
| GET | `/health/live` | public |
| GET | `/health/ready` | public |
| GET | `/metrics` | public |
| GET | `/swagger/index.html` | public |

## Layout

```
cmd/server          composition root
internal/auth       register, login, refresh, logout
internal/user       user model and CRUD
internal/config     YAML + env
internal/database   GORM
internal/cache      Redis
internal/httpx      JSON envelope, errors, pagination
internal/middleware request ID, CORS, JWT, RBAC, metrics
internal/server     router, health, graceful HTTP server
migrations          SQL schema
```

Add a domain as `internal/<name>` (handler / service / repository) and wire it in `cmd/server/main.go` and `internal/server`.

## Config

`config.yaml` holds defaults. Override secrets and hosts with env vars:

- `DATABASE_HOST`, `DATABASE_PASSWORD`, `DATABASE_NAME`
- `REDIS_HOST`, `REDIS_PASSWORD`
- `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`
- `APP_ENVIRONMENT`, `LOG_LEVEL`, `CONFIG_PATH`

Production refuses the placeholder JWT secrets (`change-me-access` / `change-me-refresh`).

## Tests

```bash
make test
INTEGRATION=1 make test-integration   # needs migrated Postgres + Redis
```
