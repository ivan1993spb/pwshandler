package chatrooms

import (
	"errors"

	"bitbucket.org/pushkin_ivan/pool-websocket-handler"
	"github.com/gorilla/websocket"
)

// Pool represents connection pool
type Pool interface {
	IsFull() bool
	IsEmpty() bool
	AddConn(*websocket.Conn) (pwshandler.Environment, error)
	DelConn(*websocket.Conn)
	HasConn(*websocket.Conn) bool
}

// PoolFactory represents pool factory
type PoolFactory interface {
	NewPool() (Pool, error)
}

// ChatRoomsPoolManager is pool manager which may be used for simple
// chat divided on rooms (pools)
type ChatRoomsPoolManager struct {
	factory PoolFactory
	pools   []Pool // Pool storage
}

// NewChatRoomsPoolManager creates new ChatRoomsPoolManager with fixed
// max number of rooms (pools) poolLimit
func NewChatRoomsPoolManager(factory PoolFactory, poolLimit uint8,
) pwshandler.PoolManager {
	return &ChatRoomsPoolManager{factory, make([]Pool, 0, poolLimit)}
}

// Implementing pwshandler.ConnManager interface
func (pm *ChatRoomsPoolManager) AddConn(conn *websocket.Conn,
) (pwshandler.Environment, error) {
	// Try to find not full pool
	for i := range pm.pools {
		if !pm.pools[i].IsFull() {
			return pm.pools[i].AddConn(conn)
		}
	}

	// Try to create pool
	if !pm.isFull() {
		// Generate new pool
		if newPool, err := pm.factory.NewPool(); err == nil {
			// Save the pool
			pm.pools = append(pm.pools, newPool)
			// Create connection to the pool
			return newPool.AddConn(conn)
		} else {
			return nil, err
		}
	}

	return nil, errors.New("Pool manager refuses to add connection")
}

// Implementing pwshandler.ConnManager interface
func (pm *ChatRoomsPoolManager) DelConn(conn *websocket.Conn) {
	for i := range pm.pools {
		// If current pool has the connection...
		if pm.pools[i].HasConn(conn) {
			// Remove it
			pm.pools[i].DelConn(conn)

			// And now if pool is empty
			if pm.pools[i].IsEmpty() {
				// Delete pool
				pm.pools = append(pm.pools[:i], pm.pools[i+1:]...)
			}

			return
		}
	}
}

// isFull returns true if pool storage is full
func (pm *ChatRoomsPoolManager) isFull() bool {
	return len(pm.pools) == cap(pm.pools)
}
