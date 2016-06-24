// PoolHandler is handler which may be useful at the projects where websocket
// connections are divided into a groups (pools) with access to common data
package pwshandler

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

// Environment is a common data for ws connections in one pool (group)
type Environment interface{}

// PoolManager is common interface for a structures which merge websocket
// connections in a pools (groups) with access to common data
type PoolManager interface {
	// AddConn creates connection to a pool and returns environment data
	AddConn(ws *websocket.Conn) (Environment, error)
	// DelConn removes passed connection from a pool if it exists in pool
	DelConn(ws *websocket.Conn) error
}

// ConnManager contains methods for processing websocket connections with
// passed common group data
type ConnManager interface {
	// Handle handles connections using passed common environment data
	Handle(ws *websocket.Conn, data Environment) error
	// HandleError processes an errors
	HandleError(ws *websocket.Conn, err error)
}

// RequestVerifier verifies requests. It has to verify a request data such
// a passed hashes, certificates, remote addr, token, passed headers or
// something else
type RequestVerifier interface {
	Verify(ws *websocket.Conn) error
}

const _ERR_FORMAT = "%s: connection handling error: %s"

// PoolHandler returns WS handler which receives websocket requests and merges
// connection goroutines in a pools (groups) with common data. poolMgr is a
// storage of groups and connections. poolMgr divides handled connections into
// groups, stores common group data and passes common data to goroutines for
// processing ws connections. connMgr contains handler for processing of ws
// connection. connMgr gets common group and ws connection. verifier must verify
// connections. If was passed nil instead verifier connections will not verified
func PoolHandler(poolMgr PoolManager, connMgr ConnManager, verifier RequestVerifier) http.Handler {

	return websocket.Handler(func(ws *websocket.Conn) {
		var err error

		// Verify request
		if verifier != nil {
			if err = verifier.Verify(ws); err != nil {
				connMgr.HandleError(ws, fmt.Errorf(_ERR_FORMAT, ws.Request().RemoteAddr, err))
				return
			}
		}

		// Create connection to a pool and take environment data
		var data Environment
		if data, err = poolMgr.AddConn(ws); err != nil {
			connMgr.HandleError(ws, fmt.Errorf(_ERR_FORMAT, ws.Request().RemoteAddr, err))
			return
		}

		// Handle connection
		if err = connMgr.Handle(ws, data); err != nil {
			connMgr.HandleError(ws, fmt.Errorf(_ERR_FORMAT, ws.Request().RemoteAddr, err))
		}

		// Delete connection from a pool (group)
		if err = poolMgr.DelConn(ws); err != nil {
			connMgr.HandleError(ws, fmt.Errorf(_ERR_FORMAT, ws.Request().RemoteAddr, err))
		}
	})
}
