FROM alpine:3.6

RUN apk update \
    && apk upgrade

RUN apk add curl bash tree tzdata \
    && cp -r -f /usr/share/zoneinfo/Hongkong /etc/localtime

ADD tfg /usr/bin/

CMD ["tfg"]






