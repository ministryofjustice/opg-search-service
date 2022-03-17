FROM golang:1.18.0

RUN go install gotest.tools/gotestsum@latest
