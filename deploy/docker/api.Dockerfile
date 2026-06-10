FROM golang:1.26.0-alpine3.23 AS base

WORKDIR /src

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .


FROM base AS security

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/securego/gosec/v2/cmd/gosec@v2.22.8 && \
    go install golang.org/x/vuln/cmd/govulncheck@1.1.14 && \
    gosec ./... && \
    govulncheck ./...


FROM base AS test

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go test ./...


FROM base AS build

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /api \
    ./cmd/api


FROM golang:1.26.0-alpine3.23 AS dev

WORKDIR /src

RUN apk add --no-cache 'git=2.54.0-r0' && \
    go install github.com/air-verse/air@v1.61.7

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]


FROM gcr.io/distroless/static-debian12:nonroot AS prod

COPY --from=build /api /api

USER 65532:65532

EXPOSE 8080

CMD ["/api"]
