FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o api .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/api .
EXPOSE 8080
CMD ["./api"]

