FROM golang:1.21.0@sha256:b490ae1f0ece153648dd3c5d25be59a63f966b5f9e1311245c947de4506981aa as build-env

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/search-service

FROM alpine:3

RUN apk --update --no-cache add \
    ca-certificates \
    && rm -rf /var/cache/apk/*
RUN apk --no-cache add tzdata

# Patch vulnerabilities
RUN apk upgrade --no-cache libcrypto3 libssl3

COPY --from=build-env /go/bin/search-service /go/bin/search-service
RUN addgroup -S app && adduser -S -g app app \
    && chown app:app /go/bin/search-service
USER app
ENTRYPOINT ["/go/bin/search-service"]
