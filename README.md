# Drumkit Integration API

A stateless Go microservice that acts as a translation layer between an internal freight TMS and the [Turvo](https://turvo.com) external TMS API. It exposes a clean, domain-focused REST API so upstream systems never have to speak Turvo's data model directly. You can also test it through the [Postman Collection](https://drive.google.com/file/d/1bb-XwBL_Ptagb7KPktW9diiF5OuM3FFj/view?usp=sharing).

This repository is the back-end part, you can also visit the [Front-end repository](https://github.com/FelipeVel/drumkit-client).

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [API Reference](#api-reference)
- [Request & Response Shapes](#request--response-shapes)
- [Configuration](#configuration)
- [Running Locally (from source)](#running-locally-from-source)
- [Running with Docker](#running-with-docker)
- [Swagger Documentation](#swagger-documentation)
- [Testing](#testing)
- [Kubernetes / AWS EKS Deployment](#kubernetes--aws-eks-deployment)
- [Observability](#observability)

---

## Overview

```
Your TMS  ──►  POST /v1/integrations/webhooks/loads  ──►  Turvo API  (create shipment)
               GET  /v1/loads                         ──►  Turvo API  (list + enrich shipments)
```

The service:

- **Translates** your internal load model to/from Turvo's shipment model.
- **Authenticates** against Turvo's OAuth2 password-grant endpoint and caches the bearer token in memory, invalidating it automatically on `401` responses.
- **Enriches** list responses by fetching full shipment and customer details from Turvo concurrently (fan-out goroutines, fan-in via channels).
- **Registers driver contacts** in Turvo before creating a shipment, and rolls them back if the shipment creation fails.

---

## Architecture

The codebase follows a strict three-layer architecture with no circular dependencies:

```
┌─────────────────────────────────────────────────────────────┐
│  HTTP Layer  (internal/controller)                          │
│  Gin handlers — binds JSON, delegates, writes response      │
└────────────────────────┬────────────────────────────────────┘
                         │ dto.*
┌────────────────────────▼────────────────────────────────────┐
│  Service Layer  (internal/service)                          │
│  Business logic — DTO ↔ domain model mapping                │
└────────────────────────┬────────────────────────────────────┘
                         │ model.*
┌────────────────────────▼────────────────────────────────────┐
│  Repository Layer  (internal/repository)                    │
│  Turvo API client — auth, HTTP calls, response mapping      │
└─────────────────────────────────────────────────────────────┘
```

**Key design choices:**

| Concern | Approach |
|---|---|
| Dependency injection | Constructor injection — no DI framework; the compiler verifies the wiring |
| Repository abstraction | `LoadRepository` interface in `internal/repository` — swappable without touching upper layers |
| Token lifecycle | In-memory `tokenCache` with mutex; invalidated on `401`, then retried once |
| Concurrency | Fan-out goroutines + buffered result channels for shipment and customer enrichment |
| Contact rollback | Driver contacts created before a shipment are deleted if shipment creation fails |
| Structured logging | `log/slog` throughout — inbound/outbound request+response pairs with latency |
| Configuration | Environment variables (`.env` file supported via `godotenv` for local dev) |

---

## Project Structure

```
drumkit-int/
├── config/
│   └── config.go              # Env-var config loader with defaults
├── docs/                      # Auto-generated Swagger spec (do not edit)
├── internal/
│   ├── controller/
│   │   └── load_controller.go # HTTP handlers for /loads routes
│   ├── dto/
│   │   └── load_dto.go        # Wire types (request/response shapes)
│   ├── middleware/
│   │   ├── cors.go            # Configurable CORS policy
│   │   └── logger.go          # Structured inbound request/response logger
│   ├── model/
│   │   ├── load.go            # Domain model (internal representation)
│   │   └── customer.go        # Customer domain model
│   ├── repository/
│   │   ├── load_repository.go       # LoadRepository interface
│   │   ├── turvo_load_repository.go # Turvo implementation
│   │   ├── turvo_auth.go            # OAuth2 token cache
│   │   ├── turvo_mapping.go         # Turvo ↔ domain model mapping
│   │   ├── turvo_status.go          # Turvo status code constants
│   │   └── turvo_types.go           # Turvo API wire types
│   └── service/
│       ├── load_service.go          # Business logic + DTO ↔ model mapping
│       └── load_service_test.go     # Unit tests
├── k8s/
│   └── deployment.yaml        # Kubernetes manifests (Secret, Deployment, Service, Ingress)
├── Dockerfile
├── go.mod
└── main.go
```

---

## API Reference

All routes are mounted under `/v1`.

### `GET /v1/loads`

Returns all freight loads currently stored in Turvo, enriched with full shipment and customer details.

**Response `200 OK`**
```json
[
  {
    "externalTMSLoadID": "9876543",
    "freightLoadID": "FL-001",
    "status": "Dispatched",
    "customer": { ... },
    "billTo":   { ... },
    "pickup":   { ... },
    "consignee":{ ... },
    "carrier":  { ... },
    "rateData": { ... },
    "specifications": { ... },
    "totalWeight": 14500,
    "routeMiles": 312.5
  }
]
```

**Error responses**

| Code | Meaning |
|---|---|
| `500` | Turvo API unreachable, auth failure, or unexpected response |

---

### `POST /v1/integrations/webhooks/loads`

Creates a new freight load in Turvo. Registers any carrier driver contacts first; rolls back contacts if shipment creation fails.

**Request body** — `application/json`

| Field | Type | Required | Notes |
|---|---|---|---|
| `freightLoadID` | string | no | Your internal load identifier |
| `status` | string | no | Initial status string |
| `customer` | object | **yes** | `externalTMSId` required |
| `pickup` | object | **yes** | `city`, `state`, `readyTime` required |
| `consignee` | object | **yes** | `city`, `state`, `apptTime` required |
| `billTo` | object | no | Bill-to party details |
| `carrier` | object | no | Carrier + driver info |
| `rateData` | object | no | Customer and carrier rate details |
| `specifications` | object | no | Handling flags (hazmat, liftgate, etc.) |
| `totalWeight` | number | no | Total shipment weight |
| `billableWeight` | number | no | Billable weight |
| `routeMiles` | number | no | Route distance in miles |
| `inPalletCount` | int | no | Pallet count at pickup |
| `outPalletCount` | int | no | Pallet count at delivery |
| `numCommodities` | int | no | Number of commodity lines |
| `poNums` | string | no | Purchase order numbers |
| `operator` | string | no | Assigned operator name |

**Response `201 Created`**
```json
{
  "id": 9876543,
  "createdAt": "2026-04-17T14:30:00.000000000Z"
}
```

**Error responses**

| Code | Meaning |
|---|---|
| `400` | Missing required fields or malformed JSON |
| `500` | Turvo API error, auth failure, or contact registration failure |

---

### `GET /swagger/*`

Interactive Swagger UI — see [Swagger Documentation](#swagger-documentation).

---

## Request & Response Shapes

### Party object (customer, billTo)

```json
{
  "externalTMSId": "12345",
  "name": "Acme Corp",
  "addressLine1": "123 Main St",
  "city": "Chicago",
  "state": "IL",
  "zipcode": "60601",
  "country": "US",
  "contact": "Jane Doe",
  "phone": "312-555-0100",
  "email": "jane@acme.com",
  "refNumber": "REF-001"
}
```

### Stop party object (pickup, consignee)

Extends Party with scheduling fields:

```json
{
  "...all party fields...",
  "businessHours": "08:00-17:00",
  "readyTime": "2026-04-20T08:00:00Z",
  "apptTime":  "2026-04-21T14:00:00Z",
  "apptNote": "Call 30 min ahead",
  "timezone": "America/Chicago",
  "warehouseId": "WH-42",
  "mustDeliver": "2026-04-21"
}
```

### Carrier object

```json
{
  "mcNumber": "123456",
  "dotNumber": "654321",
  "name": "Swift Transport",
  "phone": "800-555-0199",
  "scac": "SWFT",
  "firstDriverName": "John Driver",
  "firstDriverPhone": "555-111-2222",
  "secondDriverName": "Jane Co-Driver",
  "secondDriverPhone": "555-333-4444",
  "externalTMSTruckId": "T-001",
  "externalTMSTrailerId": "TR-001"
}
```

### Specifications object

```json
{
  "minTempFahrenheit": 34.0,
  "maxTempFahrenheit": 40.0,
  "liftgatePickup": false,
  "liftgateDelivery": true,
  "hazmat": false,
  "oversized": false,
  "tarps": false,
  "straps": true,
  "permits": false,
  "escorts": false,
  "seal": false,
  "customBonded": false,
  "labor": false
}
```

---

## Configuration

All configuration is driven by environment variables. A `.env` file in the project root is loaded automatically when present (useful for local development).

| Variable | Default | Required | Description |
|---|---|---|---|
| `SERVER_PORT` | `:8080` | no | Address and port the server listens on |
| `TURVO_BASE_URL` | — | **yes** | Base URL of the Turvo API |
| `TURVO_API_KEY` | — | **yes** | Static API key sent as `X-Api-Key` header |
| `TURVO_CLIENT_ID` | — | **yes** | OAuth2 client ID |
| `TURVO_CLIENT_SECRET` | — | **yes** | OAuth2 client secret |
| `TURVO_USERNAME` | — | **yes** | OAuth2 username (password grant) |
| `TURVO_PASSWORD` | — | **yes** | OAuth2 password (password grant) |
| `HTTP_TIMEOUT_SEC` | `10` | no | Timeout in seconds for outbound HTTP calls |
| `CORS_ALLOWED_ORIGINS` | `*` | no | Comma-separated allowed origins. Use `*` for development only |

### Example `.env` file

```dotenv
TURVO_BASE_URL=https://api.turvo.com/v1
TURVO_API_KEY=your-api-key
TURVO_CLIENT_ID=your-client-id
TURVO_CLIENT_SECRET=your-client-secret
TURVO_USERNAME=your-username
TURVO_PASSWORD=your-password
HTTP_TIMEOUT_SEC=15
CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
```

---

## Running Locally (from source)

**Prerequisites:** Go 1.25+

```bash
# 1. Clone the repository
git clone https://github.com/FelipeVel/drumkit-int.git
cd drumkit-int

# 2. Copy and fill in the environment file
cp .env.example .env   # edit .env with your Turvo credentials

# 3. Download dependencies
go mod download

# 4. Run the server
go run .
```

The server starts on `:8080` by default.

```bash
# Verify it is running
curl http://localhost:8080/v1/loads
```

### Regenerate Swagger docs (after changing annotations)

```bash
# Install swag if you don't have it
go install github.com/swaggo/swag/cmd/swag@latest

swag init
```

---

## Running with Docker

### Build and run

```bash
# Build the image
docker build -t drumkit-int:latest .

# Run with environment variables
docker run --rm \
  -p 8080:8080 \
  -e TURVO_BASE_URL=https://api.turvo.com/v1 \
  -e TURVO_API_KEY=your-api-key \
  -e TURVO_CLIENT_ID=your-client-id \
  -e TURVO_CLIENT_SECRET=your-client-secret \
  -e TURVO_USERNAME=your-username \
  -e TURVO_PASSWORD=your-password \
  drumkit-int:latest
```

### Using an env file

```bash
docker run --rm -p 8080:8080 --env-file .env drumkit-int:latest
```

### Docker build stages

The `Dockerfile` has three stages:

| Stage | Base | Purpose |
|---|---|---|
| `deps` | `golang:1.25-alpine` | Downloads Go modules (cached layer) |
| `test` | `deps` | Runs `go test ./...` — build fails here if any test fails |
| `build` | `deps` | Compiles a static binary (`CGO_ENABLED=0`) |
| runtime | `alpine:3.21` | Minimal image with only the binary and CA certificates |

The final image contains only the compiled binary — no Go toolchain, no source code.

---

## Swagger Documentation

The API is fully documented with Swagger 2.0 annotations. The interactive UI is served at:

```
http://localhost:8080/swagger/index.html
```

From there you can browse all endpoints, inspect request/response schemas, and execute live requests directly in the browser.

The generated spec files live in `docs/` and are committed to the repository. Re-generate them after changing any `// @` annotation by running `swag init` in the project root.

---

## Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detector
go test -race ./...
```

Tests live alongside the code they cover:

| File | What it tests |
|---|---|
| `internal/service/load_service_test.go` | DTO ↔ domain model mapping, service orchestration |
| `internal/repository/turvo_mapping_test.go` | Turvo API response ↔ domain model mapping |

The service layer is tested against a mock `LoadRepository`, keeping tests fast and free of external dependencies.

---

## Kubernetes / AWS EKS Deployment

All Kubernetes manifests are in `k8s/deployment.yaml`. The file contains four resources:

| Resource | Purpose |
|---|---|
| `ServiceAccount` | Binds an IAM role via IRSA for AWS API access without static credentials |
| `Secret` | Holds Turvo credentials as base64 values (consider AWS Secrets Manager + External Secrets Operator for production) |
| `Deployment` | 2 replicas spread across AZs; pulls from ECR; all secrets injected as env vars |
| `Service` | ClusterIP backend |
| `Ingress` | Internet-facing AWS ALB (AWS Load Balancer Controller); HTTP → HTTPS redirect |

### Prerequisites

- AWS Load Balancer Controller installed in the cluster
- IRSA enabled (OIDC provider associated with your EKS cluster)
- ACM certificate for your domain

### Steps to deploy

```bash
# 1. Fill in the placeholders in k8s/deployment.yaml:
#    - <ACCOUNT_ID>       → your 12-digit AWS account ID
#    - <REGION>           → e.g. us-east-1
#    - <CERTIFICATE_ARN>  → ACM certificate ARN

# 2. Base64-encode your secrets and populate the Secret data block:
echo -n 'your-turvo-api-key' | base64

# 3. Push your image to ECR
aws ecr get-login-password --region <REGION> \
  | docker login --username AWS \
    --password-stdin <ACCOUNT_ID>.dkr.ecr.<REGION>.amazonaws.com

docker build -t drumkit-int:latest .
docker tag drumkit-int:latest \
  <ACCOUNT_ID>.dkr.ecr.<REGION>.amazonaws.com/drumkit-int:latest
docker push <ACCOUNT_ID>.dkr.ecr.<REGION>.amazonaws.com/drumkit-int:latest

# 4. Apply the manifests
kubectl apply -f k8s/deployment.yaml

# 5. Retrieve the ALB hostname
kubectl get ingress drumkit-int -n default
```

---

## Observability

The service emits structured JSON logs via Go's `log/slog` to stdout. Every inbound and outbound HTTP call produces a matched pair of log entries:

**Inbound request/response (from client to this service):**
```json
{"level":"INFO","msg":"inbound request","method":"POST","path":"/v1/integrations/webhooks/loads","client_ip":"10.0.1.5","body":"{...}"}
{"level":"INFO","msg":"inbound response","method":"POST","path":"/v1/integrations/webhooks/loads","status":201,"response_size":48,"latency_ms":312}
```

**Outbound request/response (from this service to Turvo):**
```json
{"level":"INFO","msg":"outbound request","direction":"outbound","method":"POST","url":"https://api.turvo.com/v1/shipments","body":"{...}"}
{"level":"INFO","msg":"outbound response","direction":"outbound","method":"POST","url":"https://api.turvo.com/v1/shipments","status":200,"latency_ms":245,"body":"{...}"}
```

Request and response bodies are truncated to 500 characters in logs to prevent credential or PII leakage from oversized payloads.

> **Note:** The Turvo auth endpoint logs its full request body including credentials. In production, consider adding a redaction step for the `password` field before it reaches your log aggregator.
