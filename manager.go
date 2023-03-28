package fastraft

import (
	"fmt"
	"net"
	"time"
)

type Manager struct {
	nodes []*Node
}

func NewManager() *Manager {
	manager := &Manager{}
	manager.nodes = make([]*Node, 0, 16)
	return manager
}

func (self *Manager) AddNode(node *Node) {
	self.nodes = append(self.nodes, node)
}

func (self *Manager) PrintNode() {
	for _, value := range self.nodes {
		fmt.Println(value)
	}
}

func (self *Manager) MergeNode(nodes []*Node) {
	for _, node := range nodes {
		self.AddNode(node)
	}
	//copy(self.nodes[len(self.nodes):], nodes)
}

func (self *Manager) Run(address string) error {

	//	c := make(chan os.Signal)
	//	signal.Notify(c)

	//	for {
	//		s := <-c
	//		fmt.Println("get signal:", s)
	//	}

	listen, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		conn, errAccept := listen.Accept()
		if errAccept != nil {
			continue
		}
		go self.Hander(conn)
	}

}

func (self *Manager) Hander(conn net.Conn) {
	//10秒短连接
	conn.SetReadDeadline(time.Now().Add(time.Duration(10) * time.Second))
	defer conn.Close()

	fmt.Println("xxxxxxxxxx")
}
