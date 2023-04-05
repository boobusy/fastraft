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

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type raftRpcServer struct {
	*Node
	listen net.Listener
}

// NewRaftServer
// @param node
// @return *raftRpcServer
func NewRaftServer(node *Node) *raftRpcServer {
	s := new(raftRpcServer)
	s.Node = node

	go func() {

		fmt.Println("server start : ", node.inAddress)
		rpcSer := rpc.NewServer()

		err := rpcSer.RegisterName("RpcServer", s)
		if err != nil {
			fmt.Println(err)
		}

		l, err := net.Listen("tcp", node.inAddress)
		if err != nil {
			log.Fatal(err)
		}
		s.listen = l

		// use gob. rpcSer.Accept(l)
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Print("rpc.Serve: accept:", err.Error())
				return
			}
			go rpcSer.ServeConn(conn)
		}

		/*
			//  use json
			for {
				conn, e := l.Accept()
				if e != nil {
					continue
				}
				go jsonrpc.ServeConn(conn)
			}
		*/

	}()

	return s
}

// Signal Logic
// @receiver s
// @param req
// @param res
// @return error
func (s *raftRpcServer) Signal(req *RpcRequest, res *RpcResponse) error {

	// 模拟超时
	//time.Sleep(3011 * time.Millisecond)

	/*
		defer func() {
			if err := recover(); err != nil {
				res.Error = fmt.Sprintf("%s", err)
			}
		}()

	*/

	switch true {
	case req.Ping != nil:
		res.Pong = &ResponsePong{}
		s.receivePingCallBack(req.Ping, res.Pong)

	case req.Vote != nil:
		res.Vote = new(ResponseVote)
		s.receiveVoteCallBack(req.Vote, res.Vote)

	case req.Log != nil:
		res.Log = new(ResponseLog)
		s.receiveLogCallBack(req.Log, res.Log)
	}

	return nil
}

// sendAsyncSignal
// @receiver s
// @param req
// @param nodes
// @param retry
// @param rpcCallBack
// @return time.Duration
func (s *raftRpcServer) sendAsyncSignal(req *RpcRequest, nodes []*SlaveNode, retry bool, rpcCallBack func(*SlaveNode, error, *RpcResponse)) time.Duration {

	wg := new(sync.WaitGroup)
	st := time.Now()

	for _, node := range nodes {

		if retry && node.lastReq {
			continue
		}
		wg.Add(1)
		go func(nReq *SlaveNode) {

			var res RpcResponse
			err := nReq.client.Request("RpcServer.Signal", req, &res)
			if err != nil {
				nReq.lastReq = false
			} else {
				nReq.lastReq = true
			}

			rpcCallBack(node, err, &res)
			wg.Add(-1)

		}(node)
	}
	wg.Wait()

	return time.Duration(time.Now().Sub(st).Milliseconds())
}

// sendSyncSignal
// @receiver s
// @param req
// @param node
// @param rpcCallBack
// @return time.Duration
func (s *raftRpcServer) sendSyncSignal(req *RpcRequest, node *SlaveNode, rpcCallBack func(*SlaveNode, error, *RpcResponse)) time.Duration {

	st := time.Now()
	var res RpcResponse
	err := node.client.Request("RpcServer.Signal", req, &res)

	if err != nil {
		node.lastReq = false
	} else {
		node.lastReq = true
	}

	rpcCallBack(node, err, &res)

	return time.Duration(time.Now().Sub(st).Milliseconds())
}

type raftClient struct {
	address string
	conn    *rpc.Client
}

// NewRaftClient
// @param address
// @return *raftClient
func NewRaftClient(address string) *raftClient {
	c := new(raftClient)

	addr, _ := net.ResolveTCPAddr("tcp", address)
	addr.Port += 10000

	c.address = addr.String()
	return c
}

// Conn
// @receiver c
// @return *rpc.Client
// @return error
func (c *raftClient) Conn() (*rpc.Client, error) {

	if c.conn != nil {
		return c.conn, nil
	}

	addr, _ := net.ResolveTCPAddr("tcp", c.address)
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	// timeout
	//readAndWriteTimeout := 100 * time.Millisecond
	//err := conn.SetDeadline(time.Now().Add(readAndWriteTimeout))
	//if err != nil {
	//	log.Println("SetDeadline failed:", err)
	//}

	/*
		conn, err := jsonrpc.Dial("tcp", c.address)
		if err != nil {
			log.Fatal("dialing:", err)
		}
	*/

	c.conn = rpc.NewClient(conn)
	return c.conn, nil
}

// Request
// @receiver c
// @param method
// @param req
// @param res
// @return error
func (c *raftClient) Request(method string, req *RpcRequest, res *RpcResponse) error {
	conn, err := c.Conn()

	if err != nil {
		return err
	}
	call := conn.Go(method, req, res, nil)
	//return c.Conn().Call(method, req, res)

	if call.Error != nil {
		return errors.New("call.Error " + call.Error.Error())
	}

	select {
	case <-call.Done:
	case <-time.After(10 * time.Millisecond):
		return errors.New("请求超时")
	}

	return nil
}
