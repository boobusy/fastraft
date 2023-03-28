package fastraft

import (
	"bufio"
	"log"
	"net"
	//github.com/facebookgo/inject
)

// Node
// @Description:
type Node struct {
	*raftElection
	*raftEntries
	*raftRpcServer
	*SlaveNode
	nodes        []*SlaveNode
	address      string
	inAddress    string
	status       int
	nodeQuitChan NullChan
}

// NewNode
// @param name
// @param address
// @return *Node
func NewNode(name string, address string) *Node {

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		panic(err)
	}

	inAddr := *addr
	inAddr.Port += 10000
	node := defaultNode()
	node.name = name
	node.address = addr.String()
	node.inAddress = inAddr.String()
	node.raftElection = NewElection(node)
	node.raftEntries = NewRaftEntries(node)
	node.raftRpcServer = NewRaftServer(node)
	return node
}

// defaultNode
// @return *Node
func defaultNode() *Node {
	node := new(Node)
	node.SlaveNode = new(SlaveNode)
	node.nodes = make([]*SlaveNode, 0)
	return node
}

// AddNodes
// @receiver node
// @param nodes
func (node *Node) AddNodes(nodes ...string) {

	for _, n := range nodes {
		if n == node.address {
			continue
		}

		slave := &SlaveNode{
			name:    n,
			lastReq: false,
			client:  NewRaftClient(n),
		}
		node.nodes = append(node.nodes, slave)
	}
}

// Nodes
// @receiver node
// @return []*SlaveNode
func (node *Node) Nodes() []*SlaveNode {
	return node.nodes
}

// getNewNodeReqs
// @receiver node
// @return []*SlaveNode
func (node *Node) getNewNodeReqs() []*SlaveNode {

	nodesReq := make([]*SlaveNode, len(node.nodes))
	for i, n := range node.nodes {
		nodesReq[i] = new(SlaveNode)
		nodesReq[i].lastReq = false
		nodesReq[i].client = n.client
	}
	return nodesReq
}

// getNodeByName
// @receiver node
// @param name
// @return *SlaveNode
func (node *Node) getNodeByName(name string) *SlaveNode {
	for _, n := range node.nodes {
		if n.name == name {
			return n
		}
	}
	return nil
}

// DebugAddEntries
// @receiver e
// @param data
// @param team
func (e *Node) DebugAddEntries(data any, team uint32) {
	entries := &Entries{
		Commited: false,
		Team:     team,
		Data:     data,
	}
	e.raftEntries._AddEntries(entries)
}

// halfLen
// @receiver node
// @return int32
func (node *Node) halfLen() int32 {
	return int32(len(node.nodes) / 2)
}

// Run
// @receiver node
func (node *Node) Run() {
	node.eleSignalChan <- &Signal{StartSignal, nil}
	node.logSignalChan <- &Signal{StartSignal, nil}

	l, _ := net.Listen("tcp", node.address)
	defer l.Close()

	for {

		conn, err := l.Accept()

		if err != nil {
			log.Println(err)
			continue
		}

		go func(c net.Conn) {
			defer c.Close()

			r := bufio.NewReader(c)

			for {
				bytes, _, err := r.ReadLine()
				if err != nil {
					return
				}

				node.raftEntries.AddEntries(string(bytes))
			}

		}(conn)
	}

	/*
		go func() {
			time.Sleep(5 * time.Second)
			node.SignalChan <- &Signal{QuitSignal, nil}
		}()

	*/
}
