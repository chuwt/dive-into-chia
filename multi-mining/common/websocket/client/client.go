package client

import (
	"context"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"sync"
	"time"
)

type WebSocketClient struct {
	addr string // 连接地址

	conn      *websocket.Conn
	readQueue chan []byte
	handler   func([]byte)

	readErr chan error
	init    chan struct{}
	closed  chan struct{}

	reconnectTime time.Duration

	lock sync.Mutex
	log  *zap.Logger
}

/*
dial失败后自动重连，手动close后关闭
*/
func NewWebSocketClient(addr string, handler func([]byte), log *zap.Logger) WebSocketClient {

	return WebSocketClient{
		addr:          addr,
		conn:          nil,
		readQueue:     make(chan []byte, 1024),
		handler:       handler,
		readErr:       make(chan error),
		init:          make(chan struct{}),
		closed:        make(chan struct{}),
		reconnectTime: 3 * time.Second,
		log:           log.With(zap.Namespace("websocket_client")),
	}
}

func (wsc *WebSocketClient) dial() error {
	var err error
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 5 * time.Second
	wsc.conn, _, err = dialer.DialContext(context.Background(), wsc.addr, nil)
	if err != nil {
		return err
	}
	return nil
}

func (wsc *WebSocketClient) Run(onConnect func()) {

	go func() {
		defer func() {
			close(wsc.init)
			close(wsc.readQueue)
			close(wsc.readErr)
			wsc.log.Info("connect closed")
		}()

		for {
			select {
			// 主动结束
			case <-wsc.closed:
				return
			// 连接/重连
			case <-wsc.init:
				if err := wsc.dial(); err != nil {
					// 间隔后重连
					wsc.errLog("connect failed, prepare to reconnect in 3 seconds", nil, err)
					wsc.reconnectGap()
				} else {
					wsc.log.Info("connect success")
					onConnect()
					// 连接成功后启动readLoop
					go wsc.readLoop()
				}
			case msg := <-wsc.readQueue:
				wsc.handler(msg)
			case err := <-wsc.readErr:
				// readError
				wsc.errLog("read msg error", nil, err)
				_ = wsc.conn.Close()
				wsc.errLog("connect failed, prepare to reconnect in 3 seconds", nil, err)
				wsc.reconnectGap()
			}
		}
	}()

	wsc.init <- struct{}{}
}

func (wsc *WebSocketClient) reconnectGap() {
	// todo 设置重连次数提醒
	go func() {
		t := time.NewTimer(wsc.reconnectTime)
		defer t.Stop()
		select {
		case <-t.C:
			select {
			case <-wsc.closed:
				return
			default:
				wsc.init <- struct{}{}
			}
		case <-wsc.closed:
			return
		}
	}()
}

func (wsc *WebSocketClient) Close() {
	close(wsc.closed)
}

func (wsc *WebSocketClient) SendMsg(msg []byte) error {
	wsc.lock.Lock()
	defer wsc.lock.Unlock()
	err := wsc.conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}
	return nil
}

func (wsc *WebSocketClient) readLoop() {

	var (
		msg []byte
		err error
	)

	for {
		_, msg, err = wsc.conn.ReadMessage()
		if err != nil {
			wsc.readErr <- err
			return
		}
		wsc.readQueue <- msg
	}
}

func (wsc *WebSocketClient) errLog(info string, msg []byte, err error) {
	wsc.log.Error(
		info,
		zap.String("msg", string(msg)),
		zap.Error(err))
}
