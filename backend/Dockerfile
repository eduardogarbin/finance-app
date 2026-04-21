# =============================================================================
# Stage 1: base — baixa dependencias uma vez, compartilhado entre os stages
# =============================================================================
FROM golang:1.26-alpine AS base

RUN apk add --no-cache git curl

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

# =============================================================================
# Stage 2: development — usa Air para hot reload
# =============================================================================
FROM base AS development

RUN go install github.com/air-verse/air@latest

COPY . .

CMD ["air", "-c", ".air.toml"]

# =============================================================================
# Stage 3: builder — compila o binario de producao
# =============================================================================
FROM base AS builder

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /out/server ./cmd/api

# =============================================================================
# Stage 4: production — imagem minima, apenas o binario
# =============================================================================
FROM alpine:3.20 AS production

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/server .

EXPOSE 8080

CMD ["./server"]
