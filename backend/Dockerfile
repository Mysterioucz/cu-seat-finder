# One image holding all 5 service binaries; docker-compose.yml picks one per
# container with `command:` (ADR-008).
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/ ./cmd/...

# alpine, not distroless: healthchecks need busybox wget/nc.
FROM alpine:3.22
COPY --from=build /out/ /app/
