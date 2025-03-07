FROM golang:latest as builder
ARG CGO_ENABLED=0
WORKDIR /app

COPY . .
RUN go mod download
RUN go build -o schedule cmd/schedule/main.go

FROM scratch
COPY --from=builder /app/config/config.yaml /config/config.yaml
COPY --from=builder /app/schedule /schedule

EXPOSE 8080

ENTRYPOINT ["/schedule"]