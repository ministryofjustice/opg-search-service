FROM golang:1.25 as build-env

ARG TARGETARCH
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY main.go main.go
COPY internal internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o /go/bin/search-service

FROM alpine:3

RUN apk --update --no-cache add \
    ca-certificates \
    && rm -rf /var/cache/apk/*
RUN apk --no-cache add tzdata

# Patch vulnerabilities
RUN apk upgrade --no-cache busybox libcrypto3 libssl3

COPY --from=build-env /go/bin/search-service /go/bin/search-service
RUN addgroup -S app && adduser -S -g app app \
    && chown app:app /go/bin/search-service
USER app
ENTRYPOINT ["/go/bin/search-service"]
