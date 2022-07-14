package Algorithm

import (
	"container/ring"
	"net/http"
	"strings"
)

// roundRobin is strategy of scheduling that distributes work by providing
// the same amount of work to all workers.
type roundRobin struct {
	ring         *ring.Ring
	nodes        Node.nodes
	dispatchChan chan chan Node.node
	completeChan chan *http.Response
}

func newRoundRobin(hosts []string) *roundRobin {
	nodes := Node.newNodes(hosts)
	r := ring.New(len(hosts))
	for i, node := range nodes {
		node.index = i
		r.Value = node
		r = r.Next()
	}
	rr := &roundRobin{
		ring:         r,
		nodes:        nodes,
		dispatchChan: make(chan chan Node.node),
		completeChan: make(chan *http.Response),
	}
	go rr.balance()
	return rr
}

func (rr *roundRobin) Dispatch() Node.node {
	nodeChan := make(chan Node.node)
	rr.dispatchChan <- nodeChan
	return <-nodeChan
}

func (rr *roundRobin) Complete(res *http.Response) {
	rr.completeChan <- res
}

func (rr *roundRobin) balance() {
	for {
		select {
		case nodeChan := <-rr.dispatchChan:
			nodeChan <- rr.dispatch()
		case res := <-rr.completeChan:
			rr.complete(res.Request.URL.Host)
		}
	}
}

func (rr *roundRobin) dispatch() Node.node {
	node := rr.ring.Value.(*Node.node)
	node.pending += 1
	rr.ring = rr.ring.Next()
	return *node
}

func (rr *roundRobin) complete(host string) {
	rr.ring.Do(func(value interface{}) {
		node := value.(*Node.node)
		if strings.Compare(node.host, host) == 0 {
			node.pending -= 1
		}
	})
}

func (rr *roundRobin) String() string {
	return rr.nodes.String()
}
