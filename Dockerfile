# Build-Phase (Nutzt golang:alpine mit automatischer Toolchain-Anpassung)
FROM golang:alpine AS builder

WORKDIR /app

# Erlaubt Go, fehlende/neue Toolchains automatisch herunterzuladen
ENV GOTOOLCHAIN=auto

# Dependency-Dateien kopieren
COPY go.mod go.sum ./
RUN go mod download

# Den gesamten Quellcode kopieren
COPY . .

# Binary kompilieren
RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

# Laufzeit-Phase
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

# Binary und den Ordner mit den HTML-Dateien kopieren
COPY --from=builder /app/main .
COPY --from=builder /app/public ./public

EXPOSE 7777

CMD ["./main"]