FROM golang:1.25-alpine AS deps

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# ---------------------------------------------------------------------------
# Test layer — runs the full test suite; build fails here if any test fails
# ---------------------------------------------------------------------------
FROM deps AS test

COPY . .
RUN go test ./...

# ---------------------------------------------------------------------------
# Build layer — produces the static binary
# ---------------------------------------------------------------------------
FROM deps AS build

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# ---------------------------------------------------------------------------
# Runtime layer — minimal image with only the compiled binary
# ---------------------------------------------------------------------------
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=build /app/server .

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
