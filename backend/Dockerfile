FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/clawhire-api ./cmd/clawhire-api

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app
COPY --from=builder /out/clawhire-api /usr/local/bin/clawhire-api

EXPOSE 8080
CMD ["clawhire-api"]
