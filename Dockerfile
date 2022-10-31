FROM golang:1.19 as builder
WORKDIR /go/gohangout
RUN --mount=type=bind,source=.,target=/go/gohangout go build -o /tmp/ ./...

FROM alpine:3.15

ARG TZ="Asia/Shanghai"
ENV TZ ${TZ}

RUN apk upgrade --update
RUN apk --update add tzdata
RUN ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime

COPY --from=builder /tmp/gohangout /usr/local/bin/gohangout
