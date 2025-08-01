FROM golang:1.24 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .


FROM alpine

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/main /app/main
RUN chmod +x /app/main

ENTRYPOINT ["/app/main"]