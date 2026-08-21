FROM golang:1.27-alpine AS builder

WORKDIR /booking-service
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:3.23

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /booking-service

COPY --from=builder /booking-service/server .
COPY --from=builder /booking-service/resources ./resources

EXPOSE 8080
CMD ["./server"]
