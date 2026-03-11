FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o labgrab .

FROM alpine:3.21

RUN adduser -D appuser

WORKDIR /app

COPY --from=builder /build/labgrab /app/labgrab

USER appuser

CMD ["/app/labgrab"]