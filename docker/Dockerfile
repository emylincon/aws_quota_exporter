FROM golang:alpine as builder

WORKDIR /opt

COPY . .

RUN go build -o aqe .

FROM alpine:latest

RUN apk update && apk --no-cache add ca-certificates && \
    addgroup -g 1000 exporter && \
    adduser -u 1000 -D -G exporter exporter -h /exporter

WORKDIR /exporter

COPY --from=builder /opt/aqe /usr/local/bin/aqe
COPY ./docker/aws_quota_exporter/config.yml /etc/aqe/config.yml

EXPOSE 10100

USER exporter

ENTRYPOINT [ "aqe" ]

CMD ["--prom.port=10100"]
