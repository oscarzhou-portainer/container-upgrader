FROM alpine:latest

WORKDIR /var/app

COPY container-upgrader .

CMD ["./container-upgrader"]