# Jetstream Exporter


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


