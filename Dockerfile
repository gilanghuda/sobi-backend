# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod & go.sum lalu download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Build binary (nama binary = app)
RUN go build -o main main.go

# Stage 2: Run
FROM alpine:3.19

WORKDIR /root/

# Copy binary dari builder
COPY --from=builder /app/main .

RUN ls -lah

RUN chmod +x ./main

# Set default command
CMD ["./main"]