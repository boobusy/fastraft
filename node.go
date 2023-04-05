//
// MIT License
//
// # Copyright (c) 2021 Robert boobusy
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package fastraft

type INode interface {
	Node() *Node

	AddEntries(any)
	GetEntries(any)
	UpdateEntries(any)
	DelEntries(any)
}

// Node
// @Description:
type Node struct {
	*raftElection
	*raftEntries
	*raftRpcServer
	*SlaveNode
	*Option
	nodes        []*SlaveNode
	nodeQuitChan NullChan
}

func NewNode() *Node {

	return defaultNode(options)
}

// NewNodeAddress
// @param name
// @param address
// @return *Node
func NewNodeAddress(name, address string) *Node {

	opt := *options
	opt.address = address
	node := defaultNode(options)
	node.name = name
	return node
}

// NewNodeOption
// @param name
// @param opt
// @return *Node
func NewNodeOption(opt *Option) *Node {

	return defaultNode(opt)
}

// defaultNode
// @return *Node
func defaultNode(opt *Option) *Node {
	node := new(Node)
	node.Option = opt
	node.name = opt.getName()
	node.SlaveNode = new(SlaveNode)
	node.nodes = make([]*SlaveNode, 0)
	node.raftElection = NewElection(node)
	node.raftEntries = NewRaftEntries(node)
	node.raftRpcServer = NewRaftServer(node)
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
		CommitN:  1,
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

	err := node.serverFunc(node)
	if err != nil {
		node.logger.Panic(err)
	}
}
