FROM golang:1.26.0-alpine3.23 AS build

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

FROM alpine:3.19

COPY --from=build /api /api

CMD ["/api"]
