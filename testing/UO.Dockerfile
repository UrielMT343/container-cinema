FROM golang:1.25.7-alpine3.23

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# go build -o /api ./cmd/api - first is the output as -o and then the package to build
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

CMD ["/api"]
