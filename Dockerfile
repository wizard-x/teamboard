# Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY src/frontend/package.json src/frontend/package-lock.json* ./
RUN npm ci
COPY src/frontend/ .
RUN npm run build

# Build backend
FROM golang:1.22-alpine AS backend
WORKDIR /app
COPY src/backend/go.mod src/backend/go.sum ./
RUN go mod download
COPY src/backend/ .
COPY --from=frontend /app/frontend/dist ./cmd/server/static
RUN CGO_ENABLED=0 GOOS=linux go build -o /teamboard ./cmd/server

# Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=backend /teamboard /usr/local/bin/teamboard
COPY --from=backend /app/migrations /migrations
EXPOSE 8080
CMD ["teamboard"]
