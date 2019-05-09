FROM alpine:3.8
ARG TZ="Asia/Shanghai"
ENV TZ ${TZ}
RUN apk upgrade --update && apk add bash tzdata curl && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime
RUN mkdir -p /opt/gohangout
ADD gohangout /opt/gohangout/
RUN ln -s /opt/gohangout/gohangout /usr/local/bin/gohangout
