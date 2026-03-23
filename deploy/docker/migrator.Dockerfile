FROM golang:1.26.0-alpine3.23 AS builder

RUN go install github.com/jackc/tern/v2@latest

FROM alpine:3.19

WORKDIR /migrations

COPY /deploy/migrations/. ./

COPY --from=builder /go/bin/tern /usr/local/bin/tern

CMD [ "tern", "migrate" ]
