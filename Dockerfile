FROM alpine

WORKDIR /

RUN apk add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    rm -rf /var/cache/apk/*

COPY ./app-linux-amd64 /bin/app

RUN chmod +x /bin/app

EXPOSE 5071

CMD ["app"]