FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git
WORKDIR /opt
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o gitlab-external-pipeline-trigger ./cmd/main.go

FROM alpine

COPY --from=builder /opt/gitlab-external-pipeline-trigger /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/gitlab-external-pipeline-trigger"]
