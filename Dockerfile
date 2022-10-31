FROM alpine:3.15

ARG TZ="Asia/Shanghai"
ENV TZ ${TZ}

RUN apk upgrade --update
RUN apk --update add tzdata
RUN ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime

ADD gohangout /usr/local/bin/gohangout
