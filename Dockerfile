FROM golang:1.22-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build

FROM ghcr.io/viperproject/gobra:latest
COPY --from=builder /app/gobra-playground /opt/gobra-playground/server

EXPOSE 8090
ENTRYPOINT [ "/opt/gobra-playground/server" ]
