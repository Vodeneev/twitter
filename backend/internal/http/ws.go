package http

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/coder/websocket"
)

// websocket streams live notifications and direct messages to the authenticated
// user. The connection is read-only from the client's perspective; the server
// only pushes events queued via the realtime hub.
func (a *api) websocket(w http.ResponseWriter, r *http.Request) {
	me, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "authentication required")
		return
	}

	opts := &websocket.AcceptOptions{OriginPatterns: originPatterns(a.CORSOrigins)}
	if len(opts.OriginPatterns) == 0 {
		opts.InsecureSkipVerify = true
	}

	c, err := websocket.Accept(w, r, opts)
	if err != nil {
		return
	}
	defer c.CloseNow()

	client := a.Hub.Add(me.ID)
	defer a.Hub.Remove(client)

	// We don't expect inbound frames; CloseRead drains and watches for disconnect.
	ctx := c.CloseRead(r.Context())

	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case payload, ok := <-client.Send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := c.Write(writeCtx, websocket.MessageText, payload)
			cancel()
			if err != nil {
				return
			}
		case <-ping.C:
			pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := c.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

func originPatterns(corsOrigins []string) []string {
	var out []string
	for _, o := range corsOrigins {
		if u, err := url.Parse(o); err == nil && u.Host != "" {
			out = append(out, u.Host)
		}
	}
	return out
}
