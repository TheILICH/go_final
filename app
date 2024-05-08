FROM golang:1.22.2-alpine3.19 as builder

WORKDIR /app

COPY go.mod .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp

FROM alpine:3.19

WORKDIR /app

COPY wait.sh /app/
RUN chmod +x /app/wait.sh

RUN apk --no-cache add postgresql-client

COPY --from=builder /app/myapp /app/

EXPOSE 8080

# CMD ["/app/myapp"]
