FROM golang:1.25-alpine AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/athena-api ./main.go

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app && apk add --no-cache ca-certificates
USER app
WORKDIR /app
COPY --from=builder /out/athena-api /app/athena-api
EXPOSE 8080
ENTRYPOINT ["/app/athena-api"]
