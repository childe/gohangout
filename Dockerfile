FROM 1.22.5-alpine3.20 as build

RUN apk update && apk add make

WORKDIR /gohangout
COPY . .

RUN make

FROM alpine:3.20

ARG TZ="Asia/Shanghai"
ENV TZ ${TZ}

COPY --from=build /gohangout/gohangout /usr/local/bin/
