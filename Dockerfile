FROM golang:1.19-alpine AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /aca-cli

FROM alpine:3.17.0

COPY --from=builder /aca-cli /aca-cli

ENTRYPOINT ["/aca-cli"]
