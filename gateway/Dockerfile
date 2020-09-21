FROM golang:1.15.2-alpine3.12

ADD ./gateway/config.yaml config.yaml
ADD ./gateway/main .

EXPOSE 80

CMD ["./main"]