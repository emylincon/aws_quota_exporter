FROM golang:alpine AS builder

ARG APP_VERSION="0.0.0"
ARG APP_COMMIT="latest"

WORKDIR /opt

COPY . .

RUN go build -o aqe -ldflags="-X 'main.version=${APP_VERSION}' -X 'main.date=$(date "+%Y-%m-%dT%TZ")' -X 'main.commit=${APP_COMMIT}'" .

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
