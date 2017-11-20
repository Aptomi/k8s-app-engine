FROM node:9 as builder

WORKDIR /aptomi

COPY . .
RUN make w-dep w-build


FROM nginx:mainline

COPY --from=builder /aptomi/webui/dist/static /
