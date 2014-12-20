// PoolHandler is handler which may be useful at the projects where
// websocket connections are divided into a groups (pools) with a
// common environment data
//
// Author: Pushkin Ivan <pushkin13@bk.ru>
package pwshandler

import (
	"net/http"

	"golang.org/x/net/websocket"
)

// Environment is a common data for ws connections in one pool
type Environment interface{}

// PoolManager is common interface for a structures which merge
// websocket connections in a pools with access to common environment
// data of pool
type PoolManager interface {
	// AddConn creates connection to a pool and returns environment
	// data
	AddConn(ws *websocket.Conn) (Environment, error)
	// DelConn removes passed connection from a pool if exists
	DelConn(ws *websocket.Conn) error
}

// ConnManager contains methods for processing websocket connections
// with passed environment data
type ConnManager interface {
	// Handle handles connections using passed environment data
	Handle(ws *websocket.Conn, data Environment) error
	// HandleError processes an errors
	HandleError(ws *websocket.Conn, err error)
}

// RequestVerifier verifies requests. It has to verify a request data
// such a passed hashes, certificates, remote addr, passed headers or
// something else
type RequestVerifier interface {
	Verify(ws *websocket.Conn) error
}

type errHandlingConn struct {
	addr string
	err  error
}

func (e *errHandlingConn) Error() string {
	return e.addr + ": error of connection handling: " + e.err.Error()
}

// PoolHandler returns handler which receives websocket requests and
// merges connection goroutines in a pools with common data. poolMgr
// is a storage of pools and connections, connMgr contains handlers,
// verifier must verify connections. If was passed nil instead
// verifier connections will not verified
func PoolHandler(poolMgr PoolManager, connMgr ConnManager,
	verifier RequestVerifier) http.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		var err error

		// Verify request
		if verifier != nil {
			if err = verifier.Verify(ws); err != nil {
				connMgr.HandleError(ws, &errHandlingConn{
					ws.Request().RemoteAddr, err,
				})
				return
			}
		}

		// Create connection to a pool and take envitonment data
		var data Environment
		if data, err = poolMgr.AddConn(ws); err != nil {
			connMgr.HandleError(ws, &errHandlingConn{
				ws.Request().RemoteAddr, err,
			})
			return
		}

		// Handle connection
		if err = connMgr.Handle(ws, data); err != nil {
			connMgr.HandleError(ws, &errHandlingConn{
				ws.Request().RemoteAddr, err,
			})
		}

		// Delete connection from a pool
		if err = poolMgr.DelConn(ws); err != nil {
			connMgr.HandleError(ws, &errHandlingConn{
				ws.Request().RemoteAddr, err,
			})
		}
	})
}
