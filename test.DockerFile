FROM golang:1.18.1

RUN go install gotest.tools/gotestsum@latest
