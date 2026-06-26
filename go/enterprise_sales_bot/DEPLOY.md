# Deployment & Setup Instructions

## Prerequisites

- **Go:** version 1.24 or later.
- **PostgreSQL:** version 13 or later.
- **Git:** for version control and submodule management.
- **GitHub Token:** A Personal Access Token (PAT) with `repo` permissions for autonomous PR management.

## Local Setup

1. **Clone the Repository:**
   ```bash
   git clone https://github.com/robertpelloni/enterprise_sales_bot.git
   cd enterprise_sales_bot
   ```

2. **Environment Variables:** Set up the following environment variables (or use a `.env` file):

   | Variable | Required | Description |
   |---|---|---|
   | `DATABASE_URL` | Yes | `postgres://user:password@localhost:5432/sales_bot?sslmode=disable` |
   | `GITHUB_TOKEN` | Recommended | GitHub PAT for API access (enrichment, CI, PRs) |
   | `GITHUB_REPOSITORY` | Recommended | `owner/repo` for CI tracking and AutoDev |
   | `GITHUB_WEBHOOK_SECRET` | Optional | HMAC secret for webhook verification |
   | `CRM_BASE_URL` | Optional | REST CRM API base URL |
   | `CRM_API_KEY` | Optional | REST CRM API key |
| `DEPLOY_SYNC_INTERVAL` | Optional | Duration string (e.g., `1h`, `15m`) for background sync |
| `HERMES_API_URL` | Optional | Hermes Agent API server URL (e.g., `http://172.21.116.32:8642`) for real LLM |
| `HERMES_API_KEY` | Optional | Hermes API server key (must match `API_SERVER_KEY` in Hermes `.env`) |
| `HERMES_MODEL` | Optional | Model name for Hermes (default: `free-llm`) |
| `GO_TEST_MODE` | Optional | Set to `true` to skip git operations in tests |

3. **Database Migrations:** Apply migrations using your preferred tool (e.g., `golang-migrate`):
   ```bash
   # Example using a tool that supports the migrations/ directory
   migrate -path migrations/ -database "$DATABASE_URL" up
   ```

   *Note: Ensure all 4 migrations are applied, including `000004_add_interaction_success.up.sql` for the Self-Improving Prompts feature.*

4. **Initialize Submodules:**
   ```bash
   git submodule update --init --recursive
   ```

## Building the Application

Run the provided build script:

```batch
build.bat
```

This will run integrity tests and compile the binary to `bin/sales_bot.exe`.

Alternatively, build manually:

```bash
go build -v -o bin/sales_bot ./cmd/sales_bot
```

## Self-Service Deployment Dashboard

The application includes a built-in dashboard (port 8080) for managing deployment tasks autonomously.

- **Sync Repository:** Triggers a fetch and merge from the remote origin and updates all submodules recursively, ensuring the bot is running the latest code.
- **Trigger Build:** Executes the project build process (`go build`) to recompile the system on the target environment.
- **Flag Interaction Success:** Manually mark past interactions as successful to feed the Self-Improving Prompts loop.
- **View Performance Metrics:** Real-time pipeline statistics including win rate, total leads, and outreach success counts.
- **Monitor Active PRs:** Track autonomous pull requests and their merge status.

### Automated Repository Synchronization

The bot can be configured to automatically sync with its repository using two methods:

1. **GitHub Webhooks:** Configure your repository to send push events to `http://<bot-ip>:8080/api/v1/webhook/github`. This will trigger an immediate sync and build. Webhook signatures are verified via HMAC-SHA256 if `GITHUB_WEBHOOK_SECRET` is set.

2. **Background Polling:** Set the `DEPLOY_SYNC_INTERVAL` environment variable (e.g., `1h`, `15m`) to enable periodic background synchronization.

## Running the Application

Run the provided start script:

```batch
start.bat
```

Or run directly:

```bash
go run ./cmd/sales_bot
```

### Command-Line Flags

| Flag | Description |
|---|---|
| `--reconcile` | Run branch reconciliation and exit |
| `--inventory` | Generate submodule inventory table and exit |

## Docker Deployment

### Production

```bash
docker compose up -d --build
```

- Application: `http://localhost:8080`
- PostgreSQL: `localhost:5432`

### Staging

```bash
docker compose -f docker-compose.staging.yml up -d --build
```

- Application: `http://localhost:8081`
- Separate staging database

## Staging Validation

To validate the **Self-Improving Prompts** feedback loop in a staging environment:

1. Deploy using Docker Compose:
   ```bash
   docker compose -f docker-compose.staging.yml up -d --build
   ```

2. Simulate a "Closed Won" state for a lead via the CRM mock or manual DB update.

3. Verify in the logs or dashboard that past outbound interactions are flagged with `success=true`.

4. Trigger a new outreach generation and confirm that successful examples are injected into the LLM prompt.

## CI/CD

The project uses GitHub Actions for continuous integration and automated deployment:

- **CI (`ci.yml`):** Runs on push/PR to `main` — version consistency check, integrity tests, conflict resolution tests, full test suite, and build verification.
- **CI/CD (`deploy.yml`):** A unified pipeline that manages testing, staging validation, and production deployment.
- **Tests:** Runs unit and integration tests with a PostgreSQL service.
- **Staging:** Automatically deploys to a staging environment (port 8081) on pull requests and runs smoke tests.
- **Production:** Deploys to the production environment on pushes to `main` or version tags, gated by successful tests.

## Production Deployment Checklist

Before deploying to production, ensure:

- [ ] `VERSION` and `VERSION.md` are updated to the target release version.
- [ ] Database migrations are up to date (`migrations/` directory).
- [ ] Environment variables for production are configured (see above).
- [ ] The `borg` documentation submodule is initialized and up to date.
- [ ] `GITHUB_WEBHOOK_SECRET` is set for webhook signature verification.
- [ ] `GITHUB_TOKEN` is configured for CI tracking and AutoDev PR management.

### Production Verification

After deployment, run the production smoke test to verify the system's health:

```bash
TARGET_URL="https://your-production-url.com" go run scripts/smoke_test.go
```

The smoke test verifies:
- Basic health endpoint (`/health`)
- Detailed health endpoint (`/health/detailed`) — confirms database connectivity and worker liveness

### Required Secrets

To enable automated deployment, ensure the following secrets are configured in GitHub:

- `DEPLOY_HOST`: The target server address.
- `DEPLOY_KEY`: SSH private key for server access.
- `DATABASE_URL`: Production PostgreSQL connection string.

## Known Issues

- **CRLF Test Failure:** `TestResolveConflictTheirs` fails on Windows due to `\r\n` vs `\n` line ending mismatch. Does not affect production functionality.

## Hermes LLM Integration Setup

The sales bot can route all LLM calls through a local [Hermes Agent](https://github.com/NousResearch/hermes-agent) gateway for real intent classification, response generation, and sales strategy decisions.

### Prerequisites
- Hermes Agent installed and running (`hermes gateway status`)
- API server enabled in Hermes config

### Configuration (Hermes side)

In `~/.hermes/.env`:
```env
API_SERVER_ENABLED=true
API_SERVER_PORT=8642
API_SERVER_KEY=your-secret-key
API_SERVER_HOST=0.0.0.0  # required for cross-WSL/Windows access
```

Restart the gateway after changes:
```bash
systemctl --user restart hermes-gateway
```

### Configuration (Sales Bot side)

```env
HERMES_API_URL=http://<hermes-ip>:8642
HERMES_API_KEY=your-secret-key
HERMES_MODEL=free-llm
```

### Verification

```bash
# Health check
curl http://<hermes-ip>:8642/health

# LLM test
curl -H "Authorization: Bearer your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"model":"free-llm","messages":[{"role":"user","content":"Hello"}]}' \
  http://<hermes-ip>:8642/v1/chat/completions
```

### What Changes with Hermes

| Component | Without Hermes | With Hermes |
|---|---|---|
| LLM Provider | `MockLLMProvider` (returns `[MOCK LLM RESPONSE]`) | `HermesLLMProvider` (real LLM via 200+ models) |
| Intent Classifier | `MockIntentClassifier` (keyword matching) | `LLMIntentClassifier` (real LLM classification) |
| Response Generator | Template-based with mock output | Real hyper-personalized outreach drafts |
| Sales Strategy | Hardcoded heuristics | LLM-augmented sentiment analysis |
| Model Routing | N/A | Hermes handles NVIDIA → OpenRouter → LM Studio waterfall |
