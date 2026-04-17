FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

ENV SERVER_PORT=":8080" \
    HTTP_TIMEOUT_SEC="10" \
    TURVO_BASE_URL="" \
    TURVO_API_KEY="" \
    TURVO_CLIENT_ID="" \
    TURVO_CLIENT_SECRET="" \
    TURVO_USERNAME="" \
    TURVO_PASSWORD=""

EXPOSE 8080

ENTRYPOINT ["./server"]
