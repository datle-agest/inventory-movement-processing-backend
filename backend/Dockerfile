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

RUN apk add --no-cache tzdata 

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/server ./cmd/server

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/migrations ./cmd/migrations

# Step 3: final
FROM scratch

COPY --from=builder /bin/server /server
COPY --from=builder /bin/migrations /migrations
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo 
CMD ["/server"]
