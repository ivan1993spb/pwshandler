package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ivan1993spb/pwshandler"
	"golang.org/x/net/websocket"
)

func main() {
	// Verifier verifies connections
	verifier := &Verifier{}
	// Connection manager processes connections
	connMgr := &ConnManager{}
	// Pool manager stores pools and common data of pool
	poolMgr := &ChatGroupManager{NewChatGroup, make([]*ChatGroup, 0)}

	// Create ws handler
	wshandler := pwshandler.PoolHandler(poolMgr, connMgr, verifier)

	http.Handle("/chatRooms", wshandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

//
// 2. PoolManager
//

// ChatRoomData is common group data
type ChatRoomData struct {
	messages map[string]string
}

type GroupFactory func() (*ChatGroup, error)

// ChatGroup interface represents pool with connections
type ChatGroup struct {
	conns []*websocket.Conn
	data  *ChatRoomData
}

func NewChatGroup() (*ChatGroup, error) {
	return &ChatGroup{
		make([]*websocket.Conn, 0),
		&ChatRoomData{make(map[string]string)},
	}, nil
}

// IsFull returns true if pool is full
func (cg *ChatGroup) IsFull() bool {
	return false
}

// IsEmpty returns true if pool is empty
func (cg *ChatGroup) IsEmpty() bool {
	return len(cg.conns) == 0
}

// AddConn creates connection in the pool
func (cg *ChatGroup) AddConn(ws *websocket.Conn) (pwshandler.Environment, error) {
	if cg.HasConn(ws) {
		return nil, fmt.Errorf("Cannot add connection. Connection is already exists")
	}

	cg.conns = append(cg.conns, ws)

	return &ChatRoomData{make(map[string]string)}, nil
}

// DelConn removes connection from pool
func (cg *ChatGroup) DelConn(ws *websocket.Conn) error {
	for i := range cg.conns {
		// Find connection
		if cg.conns[i] == ws {
			// Remove connection
			cg.conns = append(cg.conns[:i], cg.conns[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("Cannot delete connection: connection was not found in group")
}

// HasConn returns true if passed connection belongs to the pool
func (cg *ChatGroup) HasConn(ws *websocket.Conn) bool {
	for i := range cg.conns {
		if cg.conns[i] == ws {
			return true
		}
	}

	return false
}

// ChatGroupManager manages chat groups
type ChatGroupManager struct {
	newGroup GroupFactory
	groups   []*ChatGroup
}

// Implementing pwshandler.ConnManager interface
func (pm *ChatGroupManager) AddConn(ws *websocket.Conn) (pwshandler.Environment, error) {
	// Try to find not full group
	for i := range pm.groups {
		if !pm.groups[i].IsFull() {
			return pm.groups[i].AddConn(ws)
		}
	}

	// Creating new pool
	newGroup, err := pm.newGroup()

	if err == nil {
		// Save the pool
		pm.groups = append(pm.groups, newGroup)

		// Create connection to new pool
		return newGroup.AddConn(ws)
	}

	return nil, fmt.Errorf("Cannot create new pool: %s", err)
}

// Implementing pwshandler.ConnManager interface
func (pm *ChatGroupManager) DelConn(ws *websocket.Conn) error {
	for i := range pm.groups {
		// If current pool has the connection...
		if pm.groups[i].HasConn(ws) {
			// Remove it
			err := pm.groups[i].DelConn(ws)

			// And now if pool is empty
			if pm.groups[i].IsEmpty() {
				// Delete pool
				pm.groups = append(pm.groups[:i], pm.groups[i+1:]...)
			}

			return err
		}
	}

	return fmt.Errorf("Connection to removing was not found")
}

//
// 3. ConnManager
//

// ConnManager manages connection
type ConnManager struct{}

// Implementing pwshandler.ConnManager interface
func (m *ConnManager) Handle(ws *websocket.Conn,
	data pwshandler.Environment) error {

	if chatRoomData, ok := data.(*ChatRoomData); ok {

		// Chat messaging...

		return nil
	}

	return fmt.Errorf("Chat room data was not received")
}

// Implementing pwshandler.ConnManager interface
func (m *ConnManager) HandleError(ws *websocket.Conn, err error) {
	// Log error message
	log.Printf("error: %s\n", err)
}
