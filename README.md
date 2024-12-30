# Timeseries API Server

This is Jay's Timeseries API server. It requires a Timescale DB instance.

## Getting Started

Create a `.env` file in this directory that contains necessary environment variables. Defaults:

```
USERNAME=
PASSWORD=
API_KEY=

HOST=localhost
PORT=80
JWT_SECRET=

DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USERNAME=postgres
DATABASE_PASSWORD=postgres
DATABASE_NAME=postgres

CURRENT_STORE_TYPE=memory # Options: redis, memory
REDIS_ADDRESS=localhost:6379
REDIS_PASSWORD=
REDIS_DATABASE=0

MQTT_ADDRESS=
MQTT_USERNAME=
MQTT_PASSWORD=
```

