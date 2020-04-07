FROM alpine:3.7
WORKDIR /usr/src/app/
COPY ./assets ./assets
COPY ./config ./config
COPY ./main ./main
ENTRYPOINT ["./main"]
