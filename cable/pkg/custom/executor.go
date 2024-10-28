package custom

import (
	"fmt"

	"github.com/anycable/anycable-go/common"
	"github.com/anycable/anycable-go/node"
)

// Executor handle incoming commands and client disconnections
type Executor struct {
	node node.AppNode
}

var _ node.Executor = (*Executor)(nil)

func NewExecutor(node node.AppNode) *Executor {
	return &Executor{node: node}
}

func (ex *Executor) HandleCommand(s *node.Session, msg *common.Message) (err error) {
	// Implement custom commands handling,
	// or delegate to node

	s.Log.Debug("incoming message", "msg", msg)
	switch msg.Command {
	case "subscribe":
		_, err = ex.node.Subscribe(s, msg)
	case "unsubscribe":
		_, err = ex.node.Unsubscribe(s, msg)
	case "message":
		_, err = ex.node.Perform(s, msg)
	default:
		err = fmt.Errorf("unknown command: %s", msg.Command)
	}

	return
}

func (ex *Executor) Disconnect(s *node.Session) error {
	// Do custom cleanup here
	return ex.node.Disconnect(s)
}
