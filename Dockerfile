FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/app ./cmd/app/main.go

FROM alpine
WORKDIR /app

COPY --from=builder /bin/app ./app-bin

COPY config.yaml ./

RUN addgroup -S appgroup && adduser -S appuser -G appgroup && \
    chown -R appuser:appgroup .

USER appuser

EXPOSE 8080

ENTRYPOINT ["./app-bin"]
