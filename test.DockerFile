FROM golang:1.17.8

RUN go install gotest.tools/gotestsum@latest
