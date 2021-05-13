FROM golang:alpine as builder
WORKDIR /app
RUN apk update && apk upgrade && apk add --no-cache ca-certificates make
RUN update-ca-certificates
ADD . .
RUN make linux

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/jetstream-exporter .

CMD ["./jetstream-exporter start"]
