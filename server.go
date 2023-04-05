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
	"bufio"
	"net"
	"net/http"
)

type RaftServerFunc func(node *Node) error

func RaftHttpServer(node *Node) error {

	http.HandleFunc("/set", func(writer http.ResponseWriter, request *http.Request) {

		body := make([]byte, request.ContentLength)
		_, err := request.Body.Read(body)
		if err != nil {
			writer.Write([]byte("fail"))
		} else {
			node.raftEntries.AddEntries(string(body))
			writer.Write([]byte("success"))
		}
	})
	return http.ListenAndServe(node.address, nil)
}

func RaftTcpServer(node *Node) error {

	l, err := net.Listen("tcp", node.address)
	if err != nil {
		return err
	}
	defer l.Close()

	for {

		conn, err := l.Accept()
		if err != nil {
			node.logger.Error(err)
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
}
