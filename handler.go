package pwshandler

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Environment is a common data for connections in one pool
type Environment interface{}

// ConnPool is common interface for a structures which merge websocket
// connections in a pools with access to common environment data such
// as context for starting other goroutines or something else
type ConnPool interface {
	// AddConn creates connection to pool and returns environment data
	AddConn(*websocket.Conn) (Environment, error)
	// DelConn removes passed connection from pool if exists
	DelConn(*websocket.Conn)
}

// RequestVerifier verifies requests
type RequestVerifier interface {
	Verify(*http.Request) error
}

// ConnManager contains methods for processing websocket connections
// with passed environment data
type ConnManager interface {
	// Handle handles connection using passed environment data
	Handle(*websocket.Conn, Environment) error
	// HandleError processes an errors
	HandleError(http.ResponseWriter, *http.Request, error)
}

// PoolHandler is handler which receives websocket requests and joins
// connections in a pools with common data in order with passed pool
// manager
type PoolHandler struct {
	verifier RequestVerifier
	manager  ConnManager
	pool     ConnPool
	upgrader *websocket.Upgrader
}

// NewPoolHandler creates new pool handler with passed connection
// manager and pool manager for storage and merging connections
func NewPoolHandler(manager ConnManager, pool ConnPool,
	upgrader *websocket.Upgrader) *PoolHandler {
	return &PoolHandler{nil, manager, pool, upgrader}
}

// SetupVerifier adds request verifier
func (ph *PoolHandler) SetupVerifier(verifier RequestVerifier) {
	if verifier != nil {
		ph.verifier = verifier
	}
}

// Implementing http.Handler interface
func (ph *PoolHandler) ServeHTTP(w http.ResponseWriter,
	r *http.Request) {

	var err error
	// Verify request
	if ph.verifier != nil {
		if err = ph.verifier.Verify(r); err != nil {
			ph.manager.HandleError(w, r, err)
			return
		}
	}

	// Upgrade connection to websocket
	var conn *websocket.Conn
	if conn, err = ph.upgrader.Upgrade(w, r, nil); err != nil {
		ph.manager.HandleError(w, r, err)
		return
	}

	// Create connection to pool and get envitonment data
	var env Environment
	if env, err = ph.pool.AddConn(conn); err != nil {
		ph.manager.HandleError(w, r, err)
		return
	}

	// Handle connection
	if err = ph.manager.Handle(conn, env); err != nil {
		ph.manager.HandleError(w, r, err)
	}

	// Delete connection from pool
	ph.pool.DelConn(conn)
}
