# Spectator

This program handler the camera stream by demand.

## Producer

### Usage

To send a start event in order to initialize the stream, use as follow:

```bash
producer start
```

To send a stop event in order to stop the stream, use as follow:

```bash
producer stop
```

### Build

To build the producer use

```bash
go build -o out/producer cmd/producer/main.go
```

## Consumer

Program that receives the start/stop event

### Usage

simply run the consumer with `$PATH_TO_BINARY/consumer`

### Build

To build the consumer use

```bash
go build -o out/consumer cmd/consumer/main.go
```