# Step 1: modules
FROM golang:1.26-alpine3.23 AS modules

WORKDIR /modules

COPY go.mod go.sum ./

RUN go mod download

# Step 2: builder
FROM golang:1.26-alpine3.23 AS builder

COPY --from=modules /go/pkg /go/pkg

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/server ./cmd/server

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/migrations ./cmd/migrations

# Step 3: final
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/server /app/server
COPY --from=builder /bin/migrations /app/migrations

RUN chmod +x /app/server /app/migrations

CMD ["/app/server"]
