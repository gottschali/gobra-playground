FROM ghcr.io/viperproject/gobra:latest as gobra-provider
FROM golang:1.22-alpine
RUN apk add openjdk11
RUN apk add z3
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
COPY --from=gobra-provider /gobra/gobra.jar /gobra/gobra.jar
ENV JAVA_EXE="/usr/bin/java"
ENV GOBRA_JAR=/gobra/gobra.jar
RUN go build
RUN go test -v ./...

EXPOSE 8090
ENTRYPOINT [ "/app/gobra-playground" ]
