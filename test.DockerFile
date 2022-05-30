FROM golang:1.18.2

RUN go install gotest.tools/gotestsum@latest
