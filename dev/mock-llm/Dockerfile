FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mock-llm .

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

COPY --from=builder /app/mock-llm .
COPY --from=builder /app/templates ./templates/
COPY --from=builder /app/static ./static/

EXPOSE 8080 8081

CMD ["./mock-llm"]