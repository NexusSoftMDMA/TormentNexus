# Stage 1: Build the React/TypeScript frontend
FROM node:22-slim AS frontend-builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm install
COPY . .
RUN npm run build

# Stage 2: Build the Go backend
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
# Install build tools for CGO (required by go-sqlite3)
RUN apk add --no-cache gcc musl-dev
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download
COPY backend/ ./backend/
RUN cd backend && CGO_ENABLED=1 GOOS=linux go build -o /bobby-backend ./cmd/api/main.go

# Stage 3: Final runtime image
FROM alpine:latest
WORKDIR /app

# Install dependencies (including SQLite)
RUN apk add --no-cache ca-certificates sqlite-libs libc6-compat

# Copy frontend build artifacts
COPY --from=frontend-builder /app/dist ./dist

# Copy backend binary
COPY --from=backend-builder /bobby-backend ./bobby-backend

# Copy the database
COPY bookmarks.db ./bookmarks.db

# Expose the port (Render default is 10000)
EXPOSE 10000

# Run the unified service
CMD ["./bobby-backend"]
