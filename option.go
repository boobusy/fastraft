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

import "strings"

type Option struct {
	logger        Logger
	address       string
	rpcAddress    string
	heartbeatMs   uint
	eTimeoutMsMin uint
	eTimeoutMsMax uint
	serverFunc    RaftServerFunc
}

var options = &Option{
	logger:        DefaultLogger(),
	address:       "0.0.0.0:9135",
	rpcAddress:    "0.0.0.0:19135",
	heartbeatMs:   100,
	eTimeoutMsMin: 150,
	eTimeoutMsMax: 300,
	serverFunc:    RaftTcpServer,
}

func WithLogger(log Logger) {

	options.WithLogger(log)
}

func WithRpcAddress(address string) {

	options.WithRpcAddress(address)
}

func WithHeartbeatMs(n uint) {

	options.WithHeartbeatMs(n)
}

func WithElectionMinMs(n uint) {

	options.WithElectionMinMs(n)
}

func WithElectionMaxMs(n uint) {

	options.WithElectionMaxMs(n)
}

func WithDefaultServerAddress(address string) {

	options.WithDefaultServerAddress(address)
}

func WithServer(server RaftServerFunc) {

	options.serverFunc = server
}

func (opt *Option) WithLogger(log Logger) {

	opt.logger = log
}

func (opt *Option) WithRpcAddress(address string) {

	opt.rpcAddress = address
}

func (opt *Option) WithHeartbeatMs(n uint) {

	opt.heartbeatMs = n
}

func (opt *Option) WithElectionMinMs(n uint) {

	opt.eTimeoutMsMin = n
}

func (opt *Option) WithElectionMaxMs(n uint) {

	opt.eTimeoutMsMax = n
}

func (opt *Option) WithDefaultServerAddress(address string) {

	opt.address = address
}

func (opt *Option) WithServer(server RaftServerFunc) {
	opt.serverFunc = server
}

func (opt *Option) getName() string {

	arr := strings.Split(options.address, ":")
	return "node-" + arr[1]
}
