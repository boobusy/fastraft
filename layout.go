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
	"encoding/json"
)

type (
	RequestVote struct {
		Term         uint32
		LastLogIndex uint32
		LastLogTerm  uint32
	}

	RequestLog struct {
		Team        uint32
		NextIndex   uint32
		PrevTerm    uint32
		PrevIndex   uint32
		CommitIndex uint32
		Entries     *Entries
	}

	FirstPing struct {
		NextIndex uint32
	}

	RequestPing struct {
		Team  uint32
		Time  int64
		First *FirstPing
	}

	RpcRequest struct {
		Ping *RequestPing
		Vote *RequestVote
		Log  *RequestLog
	}

	ResponseVote struct {
		Granted int8
		Team    uint32
	}

	ResponseLog struct {
		Team       uint32
		Success    bool
		MatchIndex uint32
	}

	ResponsePong struct {
		NodeName string
		Time     int64
	}

	RpcResponse struct {
		Error string
		Pong  *ResponsePong
		Vote  *ResponseVote
		Log   *ResponseLog
	}

	SlaveNode struct {
		name     string
		lastReq  bool
		client   *raftClient
		pingTime int64
		pongTime int64
	}

	Entries struct {
		Commited bool
		CommitN  int32
		Team     uint32
		Data     any
	}

	SyncRequestLog struct {
		RLog *RequestLog
		Sync chan bool
	}
)

func (v *RequestLog) String() string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}
