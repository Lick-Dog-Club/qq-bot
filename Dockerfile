FROM alpine

WORKDIR /

COPY ./app-linux-amd64 /bin/app

EXPOSE 5071

CMD ["app"]