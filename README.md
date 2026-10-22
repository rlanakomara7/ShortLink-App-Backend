# SHORTLINK API

![Go](https://img.shields.io/badge/Go-Backend-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-Framework-008ECF?style=for-the-badge)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis&logoColor=white)
![Swagger](https://img.shields.io/badge/Swagger-Swaggo-85EA2D?style=for-the-badge&logo=swagger&logoColor=black)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)

A RESTful URL shortening API built with **Go** and **Gin** for the **ShortLink App** project. The backend provides account authentication, shortlink creation and management, public redirects, and basic account information. **PostgreSQL** stores user and link data persistently, while **Redis** caches shortlink destinations to speed up redirects. This API serves a separate React frontend.

## Technologies

- **Go** — Backend programming language (see the exact version in `go.mod`).
- **Gin** — Routing, JSON request binding, and HTTP middleware.
- **PostgreSQL 16 / pgx v5** — Persistent data storage and connection pooling.
- **Redis 7 / go-redis v9** — Temporary caching of shortlink redirect destinations.
- **JWT** — Bearer-token authentication for protected routes; newly issued tokens expire after 60 minutes.
- **bcrypt** — Password hashing.
- **Swaggo / Swagger UI** — Interactive API documentation.
- **Docker Compose** — Local PostgreSQL and Redis services.
- **Air** — Optional hot reload during backend development.

## Features

- User registration and login.
- JWT-protected endpoints.
- Create shortlinks with an optional custom slug or a generated slug.
- List active shortlinks owned by the authenticated user.
- Soft-delete owned shortlinks.
- Public HTTP `301` redirect from a short slug to its original URL.
- Redis redirect caching with a 10-minute TTL; cache invalidation on deletion.
- Retrieve the authenticated user's email and account creation date.
- Swagger documentation for API exploration.

**Scope:** This backend does not currently provide click analytics, team workspaces, QR generation, password reset, or server-side avatar uploads. The frontend stores profile photos locally in the browser.

## Getting Started

### Prerequisites

- Git.
- Go (compatible with the version in `go.mod`).
- Docker with Docker Compose.
- PostgreSQL command-line access through `docker exec` for applying the initial SQL migration.
- Optional: Air for development hot reload; Swaggo CLI to regenerate docs.

The development setup runs the Go API in **WSL**, PostgreSQL and Redis in **Docker**, and the React frontend separately on Windows. Local defaults used in this project:

| Service                     | Address                 |
| --------------------------- | ----------------------- |
| Backend API                 | `http://localhost:8080` |
| PostgreSQL host connection  | `127.0.0.1:5433`        |
| Redis host connection       | `127.0.0.1:6379`        |
| Frontend development server | `http://localhost:5173` |

### Clone the repository

```bash
git clone https://github.com/rlanakomara7/ShortLink-App-Backend.git
cd ShortLink-App-Backend
```

### Configure environment variables

Copy the template and fill in your own credentials:

```bash
cp .env.example .env
```

Use the variable names **as defined in this repository's `.env.example` and `internal/config/config.go`**. The configuration includes PostgreSQL connection settings, the API port, and `JWT_SECRET`. Docker Compose also reads its PostgreSQL variables from the environment. Make sure the PostgreSQL credentials match those configured in the database.

Example placeholders (check the actual template before copying keys):

```dotenv
POSTGRES_USER=shortlink
POSTGRES_PASSWORD=replace_with_a_strong_password
POSTGRES_DB=shortlink_db
POSTGRES_PORT=5433
JWT_SECRET=replace_with_a_long_random_secret
```

Do not commit `.env`, real credentials, or access tokens. **If a secret was previously committed to GitHub, rotate it and clean the affected Git history; adding `.env` to `.gitignore` alone is not sufficient.**

### Start PostgreSQL and Redis

From the backend root in WSL:

```bash
docker compose config --quiet
docker compose up -d db redis
```

Verify Redis:

```bash
docker exec shortlink-redis redis-cli ping
```

Expected output: `PONG`.

### Apply the initial migration

The initial schema is in `migrations/001_init.sql` and creates the `users` and `links` tables with their constraints and indexes.

For a **new, empty** development database, apply it once:

```bash
docker exec -i shortlink-postgres psql -v ON_ERROR_STOP=1 -U shortlink -d shortlink_db < migrations/001_init.sql
```

Adjust the database user, database name, and container name if you changed them. **Do not rerun the migration blindly on a database that already contains tables or data.** This project does not currently include a versioned migration runner or automatic rollback workflow.

Verify the tables:

```bash
docker exec shortlink-postgres psql -U shortlink -d shortlink_db -c '\dt'
```

### Run the backend

Install module dependencies:

```bash
go mod download
```

Run directly:

```bash
go run ./cmd/api
```

Or use Air if installed:

```bash
air
```

The application connects to PostgreSQL and Redis during startup and listens on the configured API port (typically `8080`). Keep both services running for this configuration.

## API Endpoints

| Method   | Endpoint              | Authentication | Description                                |
| -------- | --------------------- | -------------- | ------------------------------------------ |
| `POST`   | `/api/register`       | Public         | Create an account                          |
| `POST`   | `/api/login`          | Public         | Authenticate and receive a JWT             |
| `POST`   | `/api/links`          | Bearer JWT     | Create a shortlink                         |
| `GET`    | `/api/links`          | Bearer JWT     | List the current user's active shortlinks  |
| `DELETE` | `/api/links/:id`      | Bearer JWT     | Soft-delete an owned shortlink             |
| `GET`    | `/api/profile`        | Bearer JWT     | Read account email and creation date       |
| `GET`    | `/:slug`              | Public         | Redirect to an active original URL (`301`) |
| `GET`    | `/swagger/index.html` | Public         | Swagger UI                                 |

Example protected request (replace the placeholder locally):

```bash
curl -H 'Authorization: Bearer YOUR_ACCESS_TOKEN' http://localhost:8080/api/links
```

### Redis redirect cache

The public redirect handler looks up a key named `shortlink:<slug>`:

1. If the key exists, use the cached destination URL.
2. Otherwise, look up the active link in PostgreSQL and cache its destination for 10 minutes.
3. When a link is soft-deleted, invalidate its Redis key so the deleted link cannot redirect from stale cache.

Inspect a test cache entry:

```bash
docker exec shortlink-redis redis-cli GET shortlink:YOUR_SLUG
docker exec shortlink-redis redis-cli TTL shortlink:YOUR_SLUG
```

Redis is a cache, **not** the authoritative source for user or shortlink data.

## Swagger API Documentation

Once the server starts, open:

**http://localhost:8080/swagger/index.html**

If you change handler annotations or add endpoints (such as `/api/profile`), regenerate the checked-in Swagger documentation using the installed Swaggo CLI. With this project layout, the entrypoint is `cmd/api/main.go`:

```bash
swag init -g cmd/api/main.go --parseDependency --parseInternal
```

Review generated changes in `docs/` before committing. Use Swagger's Authorize feature for protected routes; never commit a real JWT.

## Development Commands

| Command                                      | Description                                  |
| -------------------------------------------- | -------------------------------------------- |
| `docker compose up -d db redis`              | Start PostgreSQL and Redis                   |
| `docker compose ps`                          | Inspect service status                       |
| `go mod download`                            | Download dependencies                        |
| `go run ./cmd/api`                           | Start API directly                           |
| `air`                                        | Run with hot reload, if installed            |
| `go build ./...`                             | Compile all Go packages                      |
| `go test ./...`                              | Run available Go tests                       |
| `docker exec shortlink-redis redis-cli ping` | Check Redis connectivity                     |
| `docker compose down`                        | Stop the stack while retaining named volumes |

**Warning:** Avoid `docker compose down -v` unless you deliberately want to remove database volumes and their data.

## Data Persistence and Security Notes

- PostgreSQL data is persisted in the Docker named volume configured in `compose.yaml`.
- Redis cache entries are temporary and have a limited TTL.
- Passwords are stored as bcrypt hashes, not plaintext.
- JWTs protect user-specific operations; the frontend currently stores the token in browser local storage.
- Account photos are browser-local and are not uploaded to this backend.
- An expired or invalid JWT requires the user to sign in again.
- Before a public deployment, review secret management, HTTPS, trusted proxy configuration, API error handling, and redirect-cache consistency.

## Related Repository

- **Frontend:** [Separate React/Vite repository (add its verified GitHub URL here).](https://github.com/rlanakomara7/ShortLink-App-Frontend.git)
