//go:build !js && !wasm
// +build !js,!wasm

package wasmws

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"nhooyr.io/websocket"
)

// WebSockListener implements net.Listener and provides connections that are
// incoming websocket connections
type WebSockListener struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	acceptCh chan net.Conn
	opts     *websocket.AcceptOptions
}

var (
	_ net.Listener = (*WebSockListener)(nil)
	_ http.Handler = (*WebSockListener)(nil)
)

// NewWebSocketListener constructs a new WebSockListener, the provided context
// is for the lifetime of the listener.
// 初始化的时候加了一个 websocket.Accept时的 opts 参数
/*
	// websocket.AcceptOptions 参数
	type AcceptOptions struct {
		Subprotocols         []string 			// 支持的子协议
		InsecureSkipVerify   bool				// 是否跳过 SSL/TLS 证书的验证
		OriginPatterns       []string			// 跨域请求设置  []string("*")表示允许所有来源 也可以[]string{"*://example.com", "http://example*"} 满足一个即可访问
		CompressionMode      CompressionMode	// 压缩模式 websocket.Compression... 有三种 详见websocket源码注释
		CompressionThreshold int				// 压缩数据的最低字节数阈值
	}
*/
func NewWebSocketListener(ctx context.Context, opts *websocket.AcceptOptions) *WebSockListener {
	ctx, cancel := context.WithCancel(ctx)
	wsl := &WebSockListener{
		ctx:       ctx,
		ctxCancel: cancel,
		acceptCh:  make(chan net.Conn, 8),
		opts:      opts,
	}
	go func() { //Close queued connections
		<-ctx.Done()
		for {
			select {
			case conn := <-wsl.acceptCh:
				conn.Close()
				continue
			default:
			}
			break
		}
	}()
	return wsl
}

// ServeHTTP is a method that is mean to be used as http.HandlerFunc to accept inbound HTTP requests
// that are websocket connections
func (wsl *WebSockListener) ServeHTTP(wtr http.ResponseWriter, req *http.Request) {
	select {
	case <-wsl.ctx.Done():
		http.Error(wtr, "503: Service is shutdown", http.StatusServiceUnavailable)
		log.Printf("WebSockListener: WARN: A websocket listener's HTTP Accept was called when shutdown!")
		return
	default:
	}

	ws, err := websocket.Accept(wtr, req, wsl.opts)
	if err != nil {
		log.Printf("WebSockListener: ERROR: Could not accept websocket from %q; Details: %s", req.RemoteAddr, err)
	}

	conn := websocket.NetConn(wsl.ctx, ws, websocket.MessageBinary)
	select {
	case wsl.acceptCh <- conn:
	case <-wsl.ctx.Done():
		ws.Close(websocket.StatusBadGateway, fmt.Sprintf("Failed to accept connection before websocket listener shutdown; Details: %s", wsl.ctx.Err()))
	case <-req.Context().Done():
		ws.Close(websocket.StatusBadGateway, fmt.Sprintf("Failed to accept connection before websocket HTTP request cancelation; Details: %s", req.Context().Err()))
	}
}

// Accept fulfills the net.Listener interface and returns net.Conn that are incoming
// websockets
func (wsl *WebSockListener) Accept() (net.Conn, error) {
	select {
	case conn := <-wsl.acceptCh:
		return conn, nil
	case <-wsl.ctx.Done():
		return nil, fmt.Errorf("Listener closed; Details: %w", wsl.ctx.Err())
	}
}

// Close closes the listener
func (wsl *WebSockListener) Close() error {
	wsl.ctxCancel()
	return nil
}

// RemoteAddr returns a dummy websocket address to satisfy net.Listener
func (wsl *WebSockListener) Addr() net.Addr {
	return wsAddr{}
}

type wsAddr struct{}

func (wsAddr) Network() string { return "websocket" }

func (wsAddr) String() string { return "websocket" }
