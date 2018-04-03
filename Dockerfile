# build stage
FROM golang:alpine as build-env

COPY ./ /go/src/github.com/scouball/gollery
WORKDIR /go/src/github.com/scouball/gollery/cmd/gollery

RUN go build -o /go/bin/gollery

# final stage
FROM alpine
WORKDIR /gollery
COPY --from=build-env /go/bin/gollery /usr/bin/
COPY --from=build-env /go/src/github.com/scouball/gollery/web /web/

RUN apk --update add imagemagick

ENTRYPOINT gollery start -p /web/

EXPOSE 8080