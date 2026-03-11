FROM golang:1.24-alpine AS builder

ENV GOTOOLCHAIN=auto

RUN apk --no-cache add git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /api ./cmd/api

FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /api /api

EXPOSE 10000

CMD ["/api"]
