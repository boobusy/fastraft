// Created by GoLand
// User: boobusy.
// Date: 2023/03/21 17:12
// Email : boobusy@gmail.com

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
