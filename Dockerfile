FROM golang:1.24-alpine3.21@sha256:7772cb5322baa875edd74705556d08f0eeca7b9c4b5367754ce3f2f00041ccee

WORKDIR /app

COPY . ./

RUN go mod download && \
   go build -o /app/isle-fcrepo-fs && \
   go clean -cache -modcache

ENTRYPOINT [ "/app/isle-fcrepo-fs"]

HEALTHCHECK CMD curl -s http://localhost:8080/healthcheck | grep -q ok
