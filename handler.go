package pwshandler

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Environment is a common data for connections in one pool
type Environment interface{}

// PoolManager is common interface for a structures which merge
// websocket connections in a pools with access to common environment
// data such as context for starting other goroutines or something
// else
type PoolManager interface {
	// AddConn creates connection to a pool and returns environment
	// data
	AddConn(*websocket.Conn) (Environment, error)
	// DelConn removes passed connection from a pool if exists
	DelConn(*websocket.Conn)
}

// ConnManager contains methods for processing websocket connections
// with passed environment data
type ConnManager interface {
	// Handle handles connection using passed environment data
	Handle(*websocket.Conn, Environment) error
	// HandleError processes an errors
	HandleError(http.ResponseWriter, *http.Request, error)
}

// RequestVerifier verifies requests
type RequestVerifier interface {
	Verify(*http.Request) error
}

// PoolHandler is handler which receives websocket requests and joins
// connections in a pools with common data in order with passed pool
// manager
type PoolHandler struct {
	poolMgr  PoolManager
	connMgr  ConnManager
	verifier RequestVerifier
	upgrader *websocket.Upgrader
}

// NewPoolHandler creates new PoolHandler with passed pool manager,
// connection manager, verifier and upgrader
func NewPoolHandler(poolMgr PoolManager, connMgr ConnManager,
	verifier RequestVerifier, upgrader *websocket.Upgrader,
) http.Handler {
	return &PoolHandler{poolMgr, connMgr, verifier, upgrader}
}

// Implementing http.Handler interface
func (ph *PoolHandler) ServeHTTP(w http.ResponseWriter,
	r *http.Request) {

	var err error
	// Verify request
	if ph.verifier != nil {
		if err = ph.verifier.Verify(r); err != nil {
			ph.connMgr.HandleError(w, r, err)
			return
		}
	}

	// Upgrade connection to websocket
	var conn *websocket.Conn
	if ph.upgrader != nil {
		conn, err = ph.upgrader.Upgrade(w, r, nil)
	} else {
		conn, err = (&websocket.Upgrader{}).Upgrade(w, r, nil)
	}
	if err != nil {
		ph.connMgr.HandleError(w, r, err)
		return
	}

	// Create connection to a pool and get envitonment data
	var env Environment
	if env, err = ph.poolMgr.AddConn(conn); err != nil {
		ph.connMgr.HandleError(w, r, err)
		return
	}

	// Handle connection
	if err = ph.connMgr.Handle(conn, env); err != nil {
		ph.connMgr.HandleError(w, r, err)
	}

	// Delete connection from a pool
	ph.poolMgr.DelConn(conn)
}
