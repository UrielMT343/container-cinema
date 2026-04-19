FROM golang:1.26.0-alpine3.23 AS builder

WORKDIR /build

RUN go install github.com/jackc/tern/v2@v2.2.0 && \
  tern --version

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /migrations

USER 65532:65532

COPY deploy/migrations/*.sql ./
COPY deploy/migrations/tern.conf ./
COPY --from=builder /go/bin/tern /usr/local/bin/tern

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["tern", "status"]

CMD ["tern", "migrate"]
