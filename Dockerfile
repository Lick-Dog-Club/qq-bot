FROM alpine

WORKDIR /

RUN apk add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY ./app-linux-amd64 /bin/app

EXPOSE 5071

CMD ["app"]