FROM alpine:3.15

ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG NO_PROXY

ENV HTTP_PROXY=${HTTP_PROXY}
ENV HTTPS_PROXY=${HTTP_PROXY}
ENV NO_PROXY=${NO_PROXY}

ARG TZ="Asia/Shanghai"
ENV TZ ${TZ}

RUN apk upgrade --update
RUN apk --update add tzdata
RUN ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime
ENV HTTP_PROXY=
ENV HTTPS_PROXY=

ADD build/gohangout /usr/local/bin/gohangout
