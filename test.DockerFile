FROM golang:1.17.7

RUN go install gotest.tools/gotestsum@latest
