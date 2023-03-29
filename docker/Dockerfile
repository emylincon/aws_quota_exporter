FROM golang:alpine as builder

WORKDIR /aws_quota_exporter

COPY . .

RUN go build .

FROM alpine

WORKDIR /aws_quota_exporter

COPY --from=builder /aws_quota_exporter/aws_quota_exporter ./

CMD ./aws_quota_exporter --prom.port=10100 --config.file=${CONFIG_FILE}
