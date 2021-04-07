package core

import (
	"go-tcp-server-agent/core/intf"
	"go-tcp-server-agent/core/middleware"
	"go-tcp-server-agent/core/model"
	"go-tcp-server-agent/core/proto/msg"
	"go-tcp-server-agent/core/service"
	"net"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"
)

type gtaMsgHandler func(*service.GtaConnection, uint16, []byte, int) error

type GtaTcpServer struct {
	conf       *model.Config
	eventQueue chan *model.EventItem
	listener   net.Listener
	logger     intf.ILogger
	conns      map[int]*service.GtaConnection
	connIds    int
	handlers   map[uint16]gtaMsgHandler
}

func NewTcpServer(cfg *model.Config) *GtaTcpServer {
	logger := middleware.NewLogger(cfg.Id, false, true)

	addr := ":" + strconv.Itoa(cfg.Port)
	ls, err := net.Listen("tcp", addr)
	if err != nil {
		logger.LogError("Server listen on port %d failed", cfg.Port)
		return nil
	}

	logger.LogInfo("Server listen on [%s]", addr)

	return &GtaTcpServer{
		conf:       cfg,
		listener:   ls,
		eventQueue: make(chan *model.EventItem, 1024),
		logger:     logger,
		conns:      make(map[int]*service.GtaConnection),
		connIds:    0,
	}
}

func (svr *GtaTcpServer) Run() {
	go svr.onConnEvent()
	go svr.accept()
	svr.initHandlers()
}

func (svr *GtaTcpServer) Shutdown() {
	if svr.listener == nil {
		return
	}

	svr.logger.LogInfo("Shutdown all connections...")
	for id, c := range svr.conns {
		c.Shutdown()
		delete(svr.conns, id)
	}

	svr.logger.LogInfo("Close listener...")
	svr.listener.Close()
}

func (svr *GtaTcpServer) accept() {
	for {
		c, err := svr.listener.Accept()
		if err != nil {
			// svr.logger.LogError("Accept error: %s", err.Error())
			svr.listener = nil
			return
		}

		connection, err := service.NewConnection(c)
		if err != nil {
			return
		}

		connection.SetParser(middleware.NewMsgParser())
		connection.SetLogger(svr.logger)
		connection.SetEventQueue(svr)
		connection.Run()
	}
}

func (svr *GtaTcpServer) Push(evt *model.EventItem) {
	if svr.eventQueue == nil {
		return
	}

	for {
		select {
		case svr.eventQueue <- evt:
			{
				return
			}
		case <-time.After(5 * time.Second):
			{
				return
			}
		}
	}
}

func (svr *GtaTcpServer) Pop() <-chan *model.EventItem {
	return svr.eventQueue
}

func (svr *GtaTcpServer) onConnEvent() {
	if svr.eventQueue == nil {
		return
	}

	for {
		evt := <-svr.Pop()
		if evt == nil {
			return
		}

		// svr.logger.LogTrace("On Socket Event: %d", evt.EventType)
		switch evt.Type {
		case service.EEvType_Connected:
			{
				svr.onConnected(evt)
			}
		case service.EEvType_Data:
			{
				svr.onData(evt)
			}
		case service.EEvType_Disconnected:
			{
				svr.onDisconnected(evt)
			}
		}
	}
}

func (svr *GtaTcpServer) onConnected(evt *model.EventItem) {
	c, ok := evt.Conn.(*service.GtaConnection)
	if !ok {
		return
	}

	svr.connIds++
	c.SetId(svr.connIds)
	svr.conns[svr.connIds] = c
}

func (svr *GtaTcpServer) onDisconnected(evt *model.EventItem) {
	c, ok := evt.Conn.(*service.GtaConnection)
	if !ok {
		return
	}

	delete(svr.conns, c.GetId())
}

func (svr *GtaTcpServer) onData(evt *model.EventItem) {
	c, ok := evt.Conn.(*service.GtaConnection)
	if !ok {
		return
	}
	recvTask, ok := evt.UserData.(*service.RecvTask)
	if !ok {
		return
	}
	svr.logger.LogTrace("Connection Id: %d, recv msg id: %d, data len: %d", c.GetId(), recvTask.MsgID, recvTask.Len)

	handler, found := svr.handlers[recvTask.MsgID]
	if !found {
		return
	}

	handler(c, recvTask.MsgID, recvTask.Data, recvTask.Len)
}

func (svr *GtaTcpServer) initHandlers() {
	svr.handlers = make(map[uint16]gtaMsgHandler)

	//Add your own handler here, invoke svr.addHandler() function
	svr.addHandler(100, svr.onHandleMyFirstReq)
}

func (svr *GtaTcpServer) addHandler(msgID uint16, handler gtaMsgHandler) {
	svr.handlers[msgID] = handler
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
