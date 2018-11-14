FROM golang:alpine AS build-env
RUN apk update && apk add git
ADD . /go/src/github.com/cgilling/plantuml-proxy
RUN cd /go/src/github.com/cgilling/plantuml-proxy && go get ./... && go install

FROM alpine
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=build-env /go/bin/plantuml-proxy /app/
ENTRYPOINT ./plantuml-proxy