FROM golang:1.18rc1

RUN go install gotest.tools/gotestsum@latest
