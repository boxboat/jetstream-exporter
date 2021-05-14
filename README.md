# Jetstream Exporter

Currently the Prometheus exporter for NATS doesn't support Jetstream. This exporter will export
the total messages and bytes for a stream. It will also export the current pending messages and 
total redelivers for a consumer.

The exporter does not currently support auth.

## Config
```
servers:
  - nats://localhost:4222

streams:
  - ORDERS

consumers:
  - name: COMPLETE
    stream: ORDERS
```


