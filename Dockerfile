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

LABEL maintainer="Elias Meireles Ferreira <https://eliasmeireles.com.br>"

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/google-drive-uploader /bin/google-drive-uploader

ENTRYPOINT ["google-drive-uploader"]
