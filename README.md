# go-tcp-server-agent
## Features
- An universal multi-layer tcp server framework.
- Universal interface `ILogger` and `IParser`, support custom middleware for logger and parser.
- Support kinds of message protocol, such as proto, thrift, json, xml etc.
- Provide event queue injection, decouple business logic and network layer.
- Seperate send and receive using channel.
- Sync connection state in multiple goroutine.

## Requirements
- Golang v1.11+ Tested (go mod suppored)
- Protoc & Protoc-gen-go
```shell
$ go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
```

## Quick start
The example instance using proto, follow the steps to run:

1. Download dependecies
```shell
$ go mod download
```

2. Generate proto message
```shell
$ ./genmsg.sh
```
It will auto convert `.pb` files under core/proto/msg to `.pb.go`

3. Build and run
```shell
$ make build
$ ./build/go-tcp-server-agent
```

Or run directly

```shell
$ make run
```

## How to use

- Main code usage

```golang
//define config info for current server.
cfg := &model.Config{
    Id:   1000,
    Host: "127.0.0.1",
    Port: 1922,
}

//create a tcp server then begin to run.
if agent := core.NewTcpServer(cfg); agent != nil {
    agent.Run()
    defer agent.Shutdown()
}
```

- Receive request and send response message
In the `core/agent.go`, focus on the `initHandlers` function. 
Add your own message handler for different message id there.

```golang
func (svr *GtaTcpServer) initHandlers() {
	svr.handlers = make(map[uint16]gtaMsgHandler)

	// Add your own msg handler here.
    //e.g.
	svr.addHandler(100, svr.onHandleMyFirstReq)
}

func (svr *GtaTcpServer) onHandleMyFirstReq(c *service.GtaConnection, msgID uint16, payload []byte, len int) error {
	req := &msg.MyFirstReq{}
	if err := proto.Unmarshal(payload, req); err != nil {
		return err
	}

	ack := &msg.MyFirstAck{
		Word: []byte(string(req.Hello) + "world"),
	}

	err := c.Send(101, ack)
	return err
}

```