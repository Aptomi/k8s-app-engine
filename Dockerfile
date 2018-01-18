FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /aptomi/

ADD aptomi .

ENTRYPOINT ["./aptomi", "server"]
