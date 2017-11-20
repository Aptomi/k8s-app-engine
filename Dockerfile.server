FROM golang:1 as builder

WORKDIR /go/src/github.com/Aptomi/aptomi

RUN curl https://glide.sh/get | sh

COPY . .
RUN make vendor build


FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /aptomi/

COPY --from=builder /go/src/github.com/Aptomi/aptomi/aptomi .

ENTRYPOINT ["./aptomi"]
