# build stage
FROM golang:alpine as build-env

RUN apk --update add make curl git
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x /usr/local/bin/dep

COPY ./ /go/src/github.com/scouball/gollery
WORKDIR /go/src/github.com/scouball/gollery

RUN make install
RUN make build

# final stage
FROM alpine
WORKDIR /gollery
COPY --from=build-env /go/bin/gollery /usr/bin/

RUN apk --update add imagemagick

ENTRYPOINT gollery start

EXPOSE 8080