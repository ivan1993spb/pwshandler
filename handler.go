package pwshandler

import (
	"net/http"

	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
)

// ConnPool is common interface for a structure which merges websocket
// connections in a pools with access to common environment data such
// as context for starting other goroutines and something else
type ConnPool interface {
	// AddConn creates connection to pool and returns environment data
	// such as context and common data for selected connection pool
	AddConn(*websocket.Conn) (context.Context, interface{}, error)
	// DelConn removes passed connection from pool if exists
	DelConn(*websocket.Conn)
}

// RequestVerifier verifies request
type RequestVerifier interface {
	Verify(*http.Request) error
}

// ConnManager contains method for processing websocket connections
// with passed context and common data
type ConnManager interface {
	// Handle handles connection using passed environment data
	Handle(context.Context, *websocket.Conn, interface{}) error
	// HandleError processes an errors
	HandleError(http.ResponseWriter, error)
}

// PoolHandler is handler which receives websocket requests and joins
// connections in a pools with common data in order of passed
// connection pool manager encapsulated common data and context
type PoolHandler struct {
	verifier RequestVerifier
	manager  ConnManager
	pool     ConnPool
	upgrader *websocket.Upgrader
}

// NewPoolHandler creates new pool handler with passed connection
// manager and connection pool for storage and merging of connections
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

// Implementing Handler interface
func (ph *PoolHandler) ServeHTTP(w http.ResponseWriter,
	r *http.Request) {

	var err error
	if ph.verifier != nil {
		if err = ph.verifier.Verify(r); err != nil {
			ph.manager.HandleError(w, err)
			return
		}
	}

	var conn *websocket.Conn
	if conn, err = ph.upgrader.Upgrade(w, r, nil); err != nil {
		ph.manager.HandleError(w, err)
		return
	}

	var (
		cxt  context.Context
		data interface{}
	)
	if cxt, data, err = ph.pool.AddConn(conn); err != nil {
		ph.manager.HandleError(w, err)
		return
	}

	if err = ph.manager.Handle(cxt, conn, data); err != nil {
		ph.manager.HandleError(w, err)
	}

	ph.pool.DelConn(conn)
}
