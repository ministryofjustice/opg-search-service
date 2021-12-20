FROM golang:1.17.5

RUN go install gotest.tools/gotestsum@latest
