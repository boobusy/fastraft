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
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

// raftElection
// @Description: Election Module
type raftElection struct {
	*Node
	term atomic.Uint32
	vote atomic.Uint32
	role RoleType

	eTimeoutMsFunc  func() time.Duration
	heartbeatMs     time.Duration
	eTimeoutMsMin   time.Duration
	eTimeoutMsMax   time.Duration
	eVoteWaitTimeMs time.Duration

	heartbeat *time.Ticker
	eTimeout  *time.Timer

	eleSignalChan chan *Signal
}

// NewElection
// @param node
// @return *raftElection
func NewElection(node *Node) *raftElection {

	e := new(raftElection)
	e.Node = node
	e.eleSignalChan = make(chan *Signal, SignalSize)
	e.reset()
	go e.start()
	return e
}

// reset
// @receiver e
func (e *raftElection) reset() {

	e.term.Store(0)
	e.vote.Store(0)
	e.role = FollowerRole

	var debug time.Duration = 2000
	e.heartbeatMs = HeartbeatMs + debug
	e.eTimeoutMsMin = ETimeOutMinMs + debug
	e.eTimeoutMsMax = ETimeOutMaxMs + debug
	e.eTimeoutMsFunc = e.eTimeoutMs
}

// SetHeartbeatMs
// @receiver e
// @param d
func (e *raftElection) SetHeartbeatMs(d time.Duration) {

	e.heartbeatMs = d

	if d > time.Millisecond {
		panic(ErrHeartbeatMs)
	}

	if d < 1 {
		panic(ErrHeartbeatGt)
	}
}

// SetETimeoutMs
// @receiver e
// @param min
// @param max
func (e *raftElection) SetETimeoutMs(min, max time.Duration) {

	e.eTimeoutMsMin = min
	e.eTimeoutMsMax = max

	if min > time.Millisecond || max > time.Millisecond {
		panic(ErrETimeoutNotMs)
	}

	if min < e.heartbeatMs || min < max {
		panic(ErrETimeoutGtHB)
	}
}

// SetDebugETimeoutFunc
// @receiver e
// @param f
func (e *raftElection) SetDebugETimeoutFunc(f func() time.Duration) {

	e.eTimeoutMsFunc = f
}

// eTimeoutMs
// @receiver e
// @return time.Duration
func (e *raftElection) eTimeoutMs() time.Duration {

	d := rand.Intn(int(e.eTimeoutMsMax-e.eTimeoutMsMin)) + int(e.eTimeoutMsMin)
	return time.Duration(d) * time.Millisecond
}

// eTimeoutReset
// @receiver e
func (e *raftElection) eTimeoutReset() {

	e.eTimeout.Reset(e.eTimeoutMsFunc())
}

// raftPing
// @receiver e
func (e *raftElection) raftPing(first bool) {

	e.debug(time.Now().Format(time.StampMilli) + " send ping")
	e.pingTime = time.Now().UnixMicro()
	req := &RpcRequest{
		Ping: &RequestPing{
			Team: e.term.Load(),
			Time: e.pingTime,
		},
	}

	if first {
		req.Ping.First = &FirstPing{NextIndex: e.nextIndex.Load()}
	}

	e.raftRpcServer.sendAsyncSignal(req, e.Nodes(), false, e.rpcPingCallBack)
}

// rpcPingCallBack
// @receiver e
// @param sNode
// @param res
func (e *raftElection) rpcPingCallBack(sNode *SlaveNode, err error, res *RpcResponse) {

	if res.Pong == nil {
		return
	}

	node := e.getNodeByName(res.Pong.NodeName)
	if node != nil {
		node.pongTime = time.Now().UnixMicro()
	}
}

// receivePingCallBack
// @receiver e
// @param ping
// @param pong
func (e *raftElection) receivePingCallBack(ping *RequestPing, pong *ResponsePong) {

	e.eTimeoutReset()
	e.eleSignalChan <- &Signal{PingSignal, ping}
	pong.NodeName = e.name
	pong.Time = time.Now().UnixMicro()
}

// raftRequestVote
// @receiver e
func (e *raftElection) raftRequestVote() {

	// only one node,check leader.
	if len(e.nodes) == 0 {
		e.eleSignalChan <- &Signal{BeLeaderSignal, nil}
		return
	}

	team := e.term.Load()

	req := &RpcRequest{
		Vote: &RequestVote{
			Term:         team,
			LastLogIndex: e.lastLogIndex,
			LastLogTerm:  e.lastLogTeam,
		},
	}

	e.debug("data: ", e.address, "team = ", req.Vote.Term, "logindex = ", req.Vote.LastLogIndex, "logteam = ", req.Vote.LastLogTerm)

	// failed retry
	nodesReq := e.getNewNodeReqs()
	for team == e.term.Load() {

		// send requestVote
		et := e.raftRpcServer.sendAsyncSignal(req, nodesReq, true, e.requestVoteCallBack)

		// check leader.
		if e.checkVoteResult(0) {
			break
		}

		// delay time
		if et < e.eVoteWaitTimeMs {
			et = e.eVoteWaitTimeMs
		}
		<-time.After(et * time.Millisecond)
	}
}

// requestVoteCallBack
// @receiver e
// @param sNode
// @param res
func (e *raftElection) requestVoteCallBack(sNode *SlaveNode, err error, res *RpcResponse) {

	if res.Vote == nil {
		return
	}

	// 判断轮数,上一轮的投票丢弃与反对票丢弃
	if res.Vote.Granted == 0 || res.Vote.Team < e.term.Load() {
		return
	}

	if e.checkVoteResult(uint32(res.Vote.Granted)) {
		e.eleSignalChan <- &Signal{BeLeaderSignal, nil}
	} else {
		e.eleSignalChan <- &Signal{BadLeaderSignal, res.Vote.Team}
	}

}

// receiveVoteCallBack
// @receiver e
// @param vote
// @param resVote
func (e *raftElection) receiveVoteCallBack(vote *RequestVote, resVote *ResponseVote) {

	// once vote
	team := e.term.Load()
	if vote.Term > team && e.term.CompareAndSwap(team, vote.Term) {

		// check logIndex & logTeam
		if vote.LastLogIndex >= e.lastLogIndex && vote.LastLogTerm >= e.lastLogTeam {
			e.eTimeoutReset()

			// consistent with leader nextIndex.
			// prevention node try.
			e.nextIndex.Store(vote.LastLogIndex + 1)

			// send VoteSignal
			e.eleSignalChan <- &Signal{AgreeVoteSignal, nil}

			resVote.Granted = 1
		}

		resVote.Team = vote.Term

	} else {
		resVote.Granted = 0
		resVote.Team = team
	}
}

// checkVoteResult
// @receiver e
// @param vote
// @return bool
func (e *raftElection) checkVoteResult(vote uint32) bool {

	e.vote.Add(vote)
	return e.vote.Load() > uint32(e.halfLen())
}

// debug log
// @receiver e
// @param msg
func (e *raftElection) debug(msg ...any) {

	var role string
	switch e.role {
	case LeaderRole:
		role = "Leader"
	case CandidateRole:
		role = "Candidate"
	case FollowerRole:
		role = "Follower"
	}
	e.logger.Debug(e.address, role, msg)
}

// election start
// @receiver e
func (e *raftElection) start() {

	if startSignal := <-e.eleSignalChan; startSignal.SignalType != StartSignal {
		panic(ErrStartSignal)
	}

	if e.heartbeatMs > e.eTimeoutMsMin {
		panic("e.heartbeatMs > e.eTimeoutMsMax ")
	}

	e.heartbeat = time.NewTicker(e.heartbeatMs * time.Millisecond)
	e.eTimeout = time.NewTimer(e.eTimeoutMsMax * time.Millisecond)
	e.eVoteWaitTimeMs = e.eTimeoutMsMin / ERequestVoteN

	// start election timeout
	e.eTimeoutReset()

	// avoid two trigger
	e.heartbeat.Stop()

	go func() {
		for {
			time.Sleep(10 * time.Second)
			fmt.Println(e.name, e.commitIndex.Load(), len(e.pEntries))

			/*
				str := fmt.Sprintf("%s %d", e.name+": ", e.commitIndex.Load())
				for _, entries := range e.pEntries {
					str += fmt.Sprintf(" %+v", entries)
				}
				fmt.Println(str)

			*/
		}

	}()

	for true {

		for e.role == LeaderRole {

			select {
			case signal := <-e.eleSignalChan:
				switch signal.SignalType {
				case BeLeaderSignal:
					e.debug("be leader")

				// 有时leader也会受到投票请求
				case AgreeVoteSignal:
					e.role = FollowerRole
					e.heartbeat.Stop()

				case SendFirstPingSignal:
					e.raftPing(true)
					e.heartbeat.Reset(e.heartbeatMs * time.Millisecond)
					e.logSignalChan <- &Signal{EntriesCheckSignal, nil}

				case PingSignal:
					ping := signal.Data.(*RequestPing)

					if ping.Team > e.term.Load() {
						e.role = FollowerRole
						e.term.Store(ping.Team)
						e.debug("leader ping, update Role")
						e.pingTime = ping.Time
						e.pongTime = time.Now().UnixMicro()

						// init nextIndex
						if ping.First != nil {
							e.nextIndex.Store(ping.First.NextIndex)
						}
						//e.logSignalChan <- &Signal{EntriesQuitSyncSignal, nil}
					}

				case QuitSignal:
					goto Quit
				}

			// start ping
			case <-e.heartbeat.C:
				go e.raftPing(false)
			}
			//e.eTimeoutReset()
		}

		for e.role == FollowerRole || e.role == CandidateRole {

			select {
			case signal := <-e.eleSignalChan:

				switch signal.SignalType {
				case RequestVoteSignal:
					e.role = CandidateRole
					e.term.Add(1)
					e.vote.Store(1)
					e.debug(time.Now().Format(time.StampMilli) + " start election")
					go e.raftRequestVote()

				case BadLeaderSignal:
					team := signal.Data.(uint32)
					if e.term.Load() < team {
						e.term.Store(team)
					}

				case BeLeaderSignal:
					e.role = LeaderRole
					e.debug(" CandidateRole BeLeader")
					e.eleSignalChan <- &Signal{SendFirstPingSignal, nil}
					// try leader exit
					/*
						go func() {
							time.Sleep(15 * time.Second)
							e.eleSignalChan <- &Signal{QuitSignal, nil}
						}()
					*/

				case PingSignal:
					ping := signal.Data.(*RequestPing)
					if ping.Team > e.term.Load() {
						e.role = FollowerRole
						e.term.Store(ping.Team)
					}

					// init nextIndex
					if ping.First != nil {
						e.nextIndex.Store(ping.First.NextIndex)
						fmt.Println(e.name+" index = ", e.nextIndex.Load())
					}

					e.debug("re leader ping")
					e.pingTime = ping.Time
					e.pongTime = time.Now().UnixMicro()

				case AgreeVoteSignal:
					e.role = FollowerRole

				case QuitSignal:
					goto Quit
				}

			case <-e.eTimeout.C:
				e.eleSignalChan <- &Signal{RequestVoteSignal, nil}
				e.eTimeoutReset()
			}
		}

	}

Quit:
	e.listen.Close()
}
