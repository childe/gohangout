FROM golang:1.18-alpine3.15 as build

RUN apk update && apk add make

WORKDIR /gohangout
COPY . .

RUN make

FROM alpine:3.15

ARG TZ="Asia/Shanghai"
ENV TZ ${TZ}

RUN apk upgrade --update
RUN apk --update add tzdata
RUN ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime

COPY --from=build /gohangout/build/gohangout /usr/local/bin/
