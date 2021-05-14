# Jetstream Exporter

Currently the Prometheus exporter for NATS doesn't support Jetstream. This exporter will export
the total messages and bytes for a stream. It will also export the current pending messages and 
total redelivers for a consumer.

The exporter does not currently support auth.

## Config
```
port: 9999

servers:
  - nats://localhost:4222

streams:
  - ORDERS

consumers:
  - name: COMPLETE
    stream: ORDERS
```

## Output

```
# HELP nats_jsz_consumer_msg_pending Current number of consumer messages
# TYPE nats_jsz_consumer_msg_pending gauge
nats_jsz_consumer_msg_pending{consumer="COMPLETE:ORDERS"} 6
# HELP nats_jsz_consumer_msg_redelivery Number of redelivers
# TYPE nats_jsz_consumer_msg_redelivery counter
nats_jsz_consumer_msg_redelivery{consumer="COMPLETE:ORDERS"} 0
# HELP nats_jsz_streams_total_bytes Total number of stream bytes
# TYPE nats_jsz_streams_total_bytes counter
nats_jsz_streams_total_bytes{stream="ORDERS"} 736
# HELP nats_jsz_streams_total_msg Total number of stream messages
# TYPE nats_jsz_streams_total_msg counter
nats_jsz_streams_total_msg{stream="ORDERS"} 14
```

## Docker
The image expects a volume with the config mounted at `/config/jetstream-exporter.yaml`.
