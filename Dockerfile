FROM golang:1.21-alpine AS builder
LABEL authors="makarkonev"

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /pr-service ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /pr-service .

EXPOSE 8080

CMD ["./pr-service"]

ENTRYPOINT ["top", "-b"]