# SHORTLINK APP API

![Go](https://img.shields.io/badge/Go-1.27.1-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-1.12.0-008ECF?style=for-the-badge)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis&logoColor=white)
![Swagger](https://img.shields.io/badge/Swagger-Swaggo-85EA2D?style=for-the-badge&logo=swagger&logoColor=black)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)

A URL shortener REST API built with **Go** and **Gin**. It provides user authentication, shortlink management, and URL redirection, using **PostgreSQL** for data storage and **Redis** for caching.

## Technologies

- **Go & Gin** — REST API.
- **PostgreSQL & pgx/v5** — Database and connection pooling.
- **Redis (go-redis/v9)** — Redirect caching; JWT blacklist for logout when the logout feature is merged.
- **JWT & bcrypt** — Authentication and password hashing.
- **Swagger** — API documentation.
- **Docker Compose** — PostgreSQL and Redis services.

## Features

- Register and login with JWT access tokens (60-minute expiry).
- Create shortlinks with automatic or custom slugs.
- View and delete your own links (soft delete).
- Redirect shortlinks with Redis caching.
- View basic account profile information.
- Logout with Redis JWT blacklist **(implemented locally; verify/push before marking as released)**.

## Getting Started

### Prerequisites

Install **Go**, **Docker**, **Git**, and **Node.js/npm** for the separate frontend project.

### 1. Clone the backend

```bash
git clone https://github.com/rlanakomara7/ShortLink-App-Backend.git
cd ShortLink-App-Backend
```

### 2. Configure environment variables

Copy `.env.example` to `.env` and complete it:

```bash
cp .env.example .env
```

```dotenv
POSTGRES_USER=shortlink
POSTGRES_PASSWORD=replace_with_a_strong_password
POSTGRES_DB=shortlink_db
POSTGRES_PORT=5433

APP_PORT=8080
DB_HOST=localhost
DB_PORT=5433
DB_USER=shortlink
DB_PASSWORD=replace_with_the_same_strong_password
DB_NAME=shortlink_db
DB_SSLMODE=disable
JWT_SECRET=replace_with_a_random_secret_of_at_least_32_characters
```

Use the **same password** for `POSTGRES_PASSWORD` and `DB_PASSWORD`. Do not commit `.env`.

### 3. Start PostgreSQL and Redis

```bash
docker compose up -d db redis
```

### 4. Database setup and migrations

On a **new database**, Docker Compose creates `shortlink_db` from `POSTGRES_DB`. The schema is maintained in separate UP/DOWN files:

```text
migrations/
├── 000001_create_users.up.sql
├── 000001_create_users.down.sql
├── 000002_create_links.up.sql
└── 000002_create_links.down.sql
```

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_users_email_unique ON users (LOWER(email));
```

The Links migration creates its table and index:

```sql
CREATE TABLE links (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_url TEXT NOT NULL,
    slug VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_slug_format CHECK (slug ~ '^[A-Za-z0-9-]{3,50}$'),
    CONSTRAINT chk_slug_reserved CHECK (
        LOWER(slug) NOT IN ('api', 'login', 'register', 'dashboard')
    )
);

CREATE INDEX idx_links_user_id ON links(user_id);
```

Run **UP in order** from the backend root (WSL/Linux shell):

```bash
docker exec -i shortlink-postgres psql -v ON_ERROR_STOP=1 -U shortlink -d shortlink_db < migrations/000001_create_users.up.sql
docker exec -i shortlink-postgres psql -v ON_ERROR_STOP=1 -U shortlink -d shortlink_db < migrations/000002_create_links.up.sql
```

Rollback runs **DOWN in reverse order** (`links` before `users`), but **DOWN deletes tables and their data**. Never run it on a database containing data you want to keep. If your database has already been initialized, do **not** run these UP commands again.

### 5. Run the backend

```bash
go mod download
go run ./cmd/api
```

API: `http://localhost:8080`  
Swagger UI: `http://localhost:8080/swagger/index.html`

### 6. Run the frontend

Clone the **separate ShortLink App Frontend repository**, then run from its root:

```bash
npm install
npm run dev
```

Vite normally serves the frontend at `http://localhost:5173`. Its `/api` proxy should point to the backend on port `8080`.

## API Endpoints

| Method | Endpoint         | Description                   | Auth |
| ------ | ---------------- | ----------------------------- | ---- |
| POST   | `/api/register`  | Register an account           | No   |
| POST   | `/api/login`     | Login and get an access token | No   |
| POST   | `/api/logout`    | Revoke current JWT via Redis  | Yes  |
| GET    | `/api/profile`   | Get current profile           | Yes  |
| POST   | `/api/links`     | Create a shortlink            | Yes  |
| GET    | `/api/links`     | List your shortlinks          | Yes  |
| DELETE | `/api/links/:id` | Soft-delete a shortlink       | Yes  |
| GET    | `/:slug`         | Redirect to original URL      | No   |

For protected endpoints, send `Authorization: Bearer <access_token>`.

## Useful Commands

```bash
go build ./...           # Check that the backend compiles
go test ./...            # Run available tests
docker compose up -d     # Start services
docker compose down      # Stop services without deleting the data volume
```

**Security:** Never commit `.env` or actual credentials. Keep the `shortlink_pgdata` Docker volume if you want to preserve database records.

## Related Project

- **Frontend:** ShortLink App Frontend [(separate repository; add the verified GitHub URL here)](https://github.com/rlanakomara7/ShortLink-App-Frontend.git).

## License

No license is declared here until a `LICENSE` file is added to the repository. If MIT is chosen, add the MIT license file first.
