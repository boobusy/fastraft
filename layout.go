// Created by GoLand
// User: boobusy.
// Date: 2023/03/25 22:08
// Email : boobusy@gmail.com

package fastraft

import (
	"encoding/json"
)

type (
	RequestVote struct {
		Term         int32
		LastLogIndex uint32
		LastLogTerm  uint32
	}

	RequestLog struct {
		Team        int32
		PrevTerm    uint32
		PrevIndex   uint32
		CommitIndex uint32
		Entries     *Entries
	}

	RequestPing struct {
		Team int32
		Time int64
	}

	RpcRequest struct {
		Ping *RequestPing
		Vote *RequestVote
		Log  *RequestLog
	}

	ResponseVote struct {
		Granted int8
		Team    int32
	}

	ResponseLog struct {
		Team       int32
		Success    int8
		MatchIndex uint32
	}

	ResponsePong struct {
		NodeName string
		Time     int64
	}

	RpcResponse struct {
		Pong *ResponsePong
		Vote *ResponseVote
		Log  *ResponseLog
	}

	SlaveNode struct {
		name     string
		lastReq  bool
		client   *raftClient
		pingTime int64
		pongTime int64
	}

	// log check use
	CheckLogNode struct {
		nextIndex  uint32
		matchIndex uint32
	}

	Entries struct {
		Commited bool
		Team     uint32
		Data     any
	}

	SyncRequestLog struct {
		RLog *RequestLog
		Sync chan int8
	}
)

func (v *RequestLog) String() string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}
