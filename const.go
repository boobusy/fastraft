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
	"math/rand"
	"time"
)

const (
	FollowerRole RoleType = 1 << iota
	CandidateRole
	LeaderRole
)

const (
	SignalSize = 1 << 10

	StartSignal SignalType = 1 << iota
	QuitSignal

	PingSignal
	SendFirstPingSignal

	BeLeaderSignal
	BadLeaderSignal

	RequestVoteSignal
	AgreeVoteSignal

	EntriesAddSignal
	EntriesCommitSignal

	EntriesCheckSignal
	EntriesQuitSyncSignal
	EntriesReceiveSignal
)

const (
	ERequestVoteN = 3
	HeartbeatMs   = 1000
	ETimeOutMinMs = 1500
	ETimeOutMaxMs = 3000
)

const (
	Uint32SubOne  = 1<<32 - 1
	Uint32PlusOne = 1
)

const (
	ErrStartSignal   = "Initial Signal Must be StartSignal."
	ErrETimeoutNotMs = "eTimeoutMsMin|eTimeoutMsMax should be milliseconds."
	ErrETimeoutGtHB  = "eTimeoutMsMin|eTimeoutMsMax must be > heartbeatMs."
	ErrHeartbeatGt   = "heartbeatMs must be > 0."
	ErrHeartbeatMs   = "heartbeatMs should be milliseconds."
)

var (
	NullSignal = struct{}{}
)

type (
	NullChan chan struct{}

	RoleType uint8

	SignalType int

	Signal struct {
		SignalType
		Data any
	}

	SyncSignal struct {
		SignalType
		Data any
		Sync NullChan
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}
