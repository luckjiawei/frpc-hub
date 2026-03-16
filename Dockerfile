# Multi-stage Dockerfile for podux
# Stage 1: Build frontend
FROM --platform=linux/amd64 node:22-alpine AS frontend-builder

WORKDIR /app

COPY site/package*.json site/pnpm-lock.yaml ./
RUN npm install -g pnpm@10 && pnpm install --frozen-lockfile

COPY site/ ./
RUN pnpm run build

# Stage 2: Build backend
FROM --platform=linux/amd64 golang:1.25-alpine AS backend-builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

COPY --from=frontend-builder /app/dist ./pb_public

ARG VERSION=dev
ARG BUILD_TIME
ARG TARGETARCH
ARG TARGETVARIANT
RUN GOARM=$(echo "${TARGETVARIANT}" | tr -d 'v') \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -trimpath \
    -ldflags "-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o podux \
    main.go

# Stage 3: Final runtime image
FROM alpine:latest

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

RUN addgroup -g 1000 podux && \
    adduser -D -u 1000 -G podux podux

WORKDIR /app

COPY --from=backend-builder /app/podux .

RUN mkdir -p /app/pb_data && \
    chown -R podux:podux /app

USER podux

EXPOSE 8090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/api/health || exit 1

ENTRYPOINT ["/app/podux"]
CMD ["serve", "--http", "0.0.0.0:8090"]
