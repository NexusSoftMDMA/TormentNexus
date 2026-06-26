# Deployment Instructions

## Environment Setup
The Borg Intelligence Harvester is intended to run primarily via background scripts and web interfaces.

### Python Backend
Dependencies are managed via `pip`:
```bash
pip install -r requirements.txt
```
To run the research worker:
```bash
python3 deep_research.py
```

### Go Backend API
To build and verify the Go API backend:
```bash
cd backend
go build -buildvcs=false ./cmd/api
go test -buildvcs=false ./internal/...
```

### Static Website Landing Pages
The frontend landing pages (`tormentnexus.site` and `hypernexus.site`) are deployed using Cloudflare Pages via GitHub Actions.
Any commits to the `landing/` directory on the `main` branch will automatically trigger deployment.

## Future Restorations
Due to the loss of `catalog.db` and the corresponding skill extraction datasets, future deployments should re-index tools and recreate databases using available fallback scripts in the `scripts/` directory before starting main server operations.
