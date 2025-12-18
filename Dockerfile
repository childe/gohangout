FROM golang:1.25.5-alpine3.23 as build

RUN apk update && apk add make git

WORKDIR /gohangout
COPY . .

RUN make

FROM alpine:3.23

ARG TZ="Asia/Shanghai"
ENV TZ ${TZ}

COPY --from=build /gohangout/gohangout /usr/local/bin/
