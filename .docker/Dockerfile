# Step 1: Modules caching
FROM golang:1.24-alpine3.20 AS modules

ENV \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /modules

COPY go.mod go.sum ./

RUN go mod download

# Step 2: Build a special service
FROM golang:1.24-alpine3.20 AS builder

ARG APP_NAME

RUN \
    apk update && \
    apk add git

COPY --from=modules /go/pkg /go/pkg
COPY . /app

WORKDIR /app

RUN go build --buildvcs=true -o /bin/app ./cmd/$APP_NAME

# Step 3: Reduce the size as much as possible
FROM scratch

COPY --from=builder /bin/app /app
COPY --from=modules /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/config/config.yaml /config/config.yaml

CMD ["/app"]
