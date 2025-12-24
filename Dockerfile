# Use the official Golang image
FROM golang:1.25-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

COPY . .

# Download dependencies
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o google-drive-uploader cmd/uploader/main.go

FROM alpine:latest

LABEL org.opencontainers.image.title="Google Drive Uploader"
LABEL org.opencontainers.image.description="Google Drive Uploader"
LABEL org.opencontainers.image.version="0.0.1"
LABEL org.opencontainers.image.authors="Elias Ferreira <https://github.com/eliasmeireles>"
LABEL org.opencontainers.image.source="https://github.com/eliasmeireles/google-drive-uploader"
LABEL org.opencontainers.image.licenses="MIT"
LABEL maintainer="Elias Ferreira <https://eliasmeireles.com.br>"

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/google-drive-uploader /bin/google-drive-uploader

ENTRYPOINT ["google-drive-uploader"]
