FROM golang:1.23-alpine AS build

ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

ARG CGO=1
ENV CGO_ENABLED=${CGO}
ENV GOOS=linux
ENV GO111MODULE=on

WORKDIR /go/src/github.com/pedromol/bashhub-server

COPY . /go/src/github.com/pedromol/bashhub-server/

RUN apk update && \
    apk add --no-cache g++ gcc musl-dev


RUN go build \
    cmd/bashhub-server/main.go \
    -ldflags "-X github.com/pedromol/bashhub-server/cmd.Version=${VERSION} -X github.com/pedromol/bashhub-server/cmd.GitCommit=${GIT_COMMIT} -X github.com/pedromol/bashhub-server/cmd.BuildDate=${BUILD_DATE}" \
    -o /go/bin/bashhub-server

# ---

FROM alpine:3

COPY --from=build /go/bin/bashhub-server /usr/bin/bashhub-server

RUN apk update && \
    apk add --no-cache ca-certificates libc6-compat libstdc++

VOLUME /data
WORKDIR /data

EXPOSE 8080

ENTRYPOINT ["bashhub-server"]
