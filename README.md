# GitHub Release Notification API

This service allows users to subscribe to email notifications about new releases of a GitHub repository.

The project is implemented as a single Go monolith and includes:

- an HTTP API for subscription management
- a background scanner for checking new releases
- a notifier for sending emails
- PostgreSQL for storing all data
- automatic migration execution on service startup

The Swagger contract is available in the [swagger.yaml](https://mykhailo-hrynko.github.io/se-school/task/swagger.yaml).

## Main features

- `POST /api/subscribe` creates a subscription for a repository (format: `owner/repo`)
- `GET /api/confirm/{token}` confirms an email subscription
- `GET /api/unsubscribe/{token}` unsubscribes from release notifications
- `GET /api/subscriptions?email=...` returns all active subscriptions
- when a subscription is created, the repository is validated through the GitHub API
- `last_seen_tag` is stored for each repository
- the scanner regularly checks only active and confirmed subscriptions
- an email is sent only when a new release appears

## How it works

1. The user calls `POST /api/subscribe` with `email` and `repo`.
2. The service validates the email and the repository format `owner/repo`.
3. The service checks whether the repository exists through the GitHub API.
4. If the repository is not yet stored in the database, a record is created in the `repositories` table.
5. During the first save, the service reads the current latest release tag and stores it in `last_seen_tag` so that it does not send an email about an old release.
6. A subscription is created with a `confirm_token` and an `unsubscribe_token`.
7. A confirmation email is sent to the user.
8. After confirmation, the background scanner starts taking this subscription into account.
9. If a new tag is found, the service sends emails to all active confirmed subscribers and updates `last_seen_tag`.

## Architecture

The project is a monolith and contains three logical parts within a single process:

- API: a Gin router with subscription endpoints
- Scanner: a background scheduler that calls `ScanOnce()` at the `SCAN_INTERVAL`
- Notifier: an SMTP email sender

The scanner runs in a goroutine after the application starts. Migrations are executed automatically during the startup of `cmd/app/main.go`.

## Stack

- Go 1.26
- Gin
- PostgreSQL
- `golang-migrate` for migrations
- `gomail` for SMTP
- Docker / Docker Compose

## Database

Migrations are located in the [migrations](./migrations) directory.

The current schema includes two main tables:

- `repositories`
    - `full_name`
    - `owner`
    - `name`
    - `last_seen_tag`
    - `last_checked_at`
    - `created_at`
    - `updated_at`
- `subscriptions`
    - `email`
    - `repository_id`
    - `confirmed`
    - `active`
    - `confirm_token`
    - `unsubscribe_token`
    - `confirmed_at`
    - `created_at`
    - `updated_at`

Migrations are executed automatically on startup using `golang-migrate`.

## Scanner/notifier behavior

- the scanner works only with active and confirmed subscriptions
- if a repository has no releases, `last_seen_tag` may be `NULL`
- if the latest release tag has not changed, no emails are sent
- if sending at least one email fails, `last_seen_tag` is not updated
- `last_checked_at` is updated after a successful request to the GitHub API

## Tests

The project includes unit tests for the business logic of services and handlers.

Run:

```bash
go test ./...
```

## Local run

### 1. Prepare dependencies

You need:

- Go 1.26
- PostgreSQL 16+
- an SMTP server or mail catcher

### 2. Configure `.env`

Update the values in `.env` for your environment.

Recommendation: do not store real SMTP passwords or GitHub tokens in the repository. For demonstration, it is better to use a separate test account or a local SMTP service.

### 3. Run the application

```bash
go run ./cmd/app
```

When starting, the service:

- reads the configuration
- connects to PostgreSQL
- runs migrations
- starts the background scheduler
- starts the HTTP server on `APP_PORT`

## Run with Docker Compose

The repository includes:

- [Dockerfile](./Dockerfile)
- [docker-compose.yml](./docker-compose.yml)

Run command:

```bash
docker compose up --build
```

What is currently started:

- `postgres`
- `app`

The service will be available at:

```text
http://localhost:8080
```

## Deployment

The application is deployed on Render:

- Web Service (Go API)
- PostgreSQL (managed database)

Key deployment details:

- environment variables are configured in Render
- the application reads the port from the PORT environment variable in production
- /health endpoint is used for health checks
- HTML templates are embedded into the binary to ensure correct behavior in production

### HTML subscription page

In addition to the REST API, the project includes a simple HTML page for subscribing to GitHub release notifications.

- GET /subscribe returns an HTML form
- the form allows entering:
   - email
   - repository in owner/repo format
- after submission, the page sends a request to POST /api/subscribe
- the result (success or error) is displayed directly on the page

### Public page:
https://github-release-notification-api-qj7p.onrender.com/subscribe