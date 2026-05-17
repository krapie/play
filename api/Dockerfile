FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o play-api .

FROM alpine:3.19 AS runner
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/play-api .
EXPOSE 8080
CMD ["./play-api"]
