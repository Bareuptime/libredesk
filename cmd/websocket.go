package main

import (
	"fmt"

	amodels "github.com/abhinavxd/libredesk/internal/auth/models"
	"github.com/abhinavxd/libredesk/internal/ws"
	wsmodels "github.com/abhinavxd/libredesk/internal/ws/models"
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

// ErrHandler is a custom error handler.
func ErrHandler(ctx *fasthttp.RequestCtx, status int, reason error) {
	fmt.Printf("error status %d: %s", status, reason)
}

// upgrader is a websocket upgrader.
var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
	},
	Error: ErrHandler,
}

// handleWS handles the websocket connection.
func handleWS(r *fastglue.Request, hub *ws.Hub) error {
	var (
		auser = r.RequestCtx.UserValue("user").(amodels.User)
		app   = r.Context.(*App)
	)
	err := upgrader.Upgrade(r.RequestCtx, func(conn *websocket.Conn) {
		c := ws.Client{
			ID:   auser.ID,
			Hub:  hub,
			Conn: conn,
			Send: make(chan wsmodels.WSMessage, 10000),
		}
		hub.AddClient(&c)
		go c.Listen()
		c.Serve()
	})
	if err != nil {
		app.lo.Error("error upgrading tcp connection", "user_id", auser.ID, "error", err)
	}
	return nil
}
