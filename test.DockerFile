FROM golang:1.17.6

RUN go install gotest.tools/gotestsum@latest
