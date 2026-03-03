FROM golang:alpine AS builder

WORKDIR /app

# Copy module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o ytclone .

FROM alpine:latest

# Install dependencies: ffmpeg (for merging audio/video) and curl (to download yt-dlp)
RUN apk update && \
    apk add --no-cache ffmpeg curl python3 && \
    curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

WORKDIR /app

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/ytclone .

# Create a temp directory for downloads
RUN mkdir -p temp

# The application requires an interactive terminal for the first Google OAuth run
ENTRYPOINT ["./ytclone"]
