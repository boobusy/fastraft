package fastraft

import (
	"log"
	"math/rand"
	"sync/atomic"
	"time"
)

// raftElection
// @Description: Election Module
type raftElection struct {
	*Node
	term atomic.Int32
	vote atomic.Int32
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
func (e *raftElection) raftPing() {

	e.log(time.Now().Format(time.StampMilli) + " send ping")
	e.pingTime = time.Now().UnixMicro()
	req := &RpcRequest{
		Ping: &RequestPing{
			Team: e.term.Load(),
			Time: e.pingTime,
		},
	}

	e.raftRpcServer.sendSignal(req, e.Nodes(), false, e.rpcPingCallBack)
}

// rpcPingCallBack
// @receiver e
// @param sNode
// @param res
func (e *raftElection) rpcPingCallBack(sNode *SlaveNode, res *RpcResponse) {

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

	team := e.term.Load()

	req := &RpcRequest{
		Vote: &RequestVote{
			Term:         team,
			LastLogIndex: e.lastLogIndex,
			LastLogTerm:  e.lastLogTeam,
		},
	}

	e.log("data: ", e.address, "team = ", req.Vote.Term, "logindex = ", req.Vote.LastLogIndex, "logteam = ", req.Vote.LastLogTerm)

	// failed retry
	nodesReq := e.getNewNodeReqs()
	for team == e.term.Load() {

		// send requestVote
		et := e.raftRpcServer.sendSignal(req, nodesReq, true, e.requestVoteCallBack)

		// check leader
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
func (e *raftElection) requestVoteCallBack(sNode *SlaveNode, res *RpcResponse) {

	if res.Vote == nil {
		return
	}

	// 判断轮数,上一轮的投票丢弃与反对票丢弃
	if res.Vote.Granted == 0 || res.Vote.Team < e.term.Load() {
		return
	}

	if e.checkVoteResult(int32(res.Vote.Granted)) {
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
	if vote.Term > e.term.Load() {
		e.term.Store(vote.Term)

		// check logindex & logteam
		if vote.LastLogIndex >= e.lastLogIndex && vote.LastLogTerm >= e.lastLogTeam {
			e.eTimeoutReset()

			// 赞成票就和leader的nextIndex保持一致
			e.nextIndex.Store(vote.LastLogIndex + 1)

			resVote.Granted = 1
			e.eleSignalChan <- &Signal{AgreeVoteSignal, nil}

		} else {
			resVote.Granted = 0
			e.log("granted flase")
		}
	}

	resVote.Team = e.term.Load()
}

// checkVoteResult
// @receiver e
// @param vote
// @return bool
func (e *raftElection) checkVoteResult(vote int32) bool {

	e.vote.Add(vote)
	return e.vote.Load() > e.halfLen()
}

// log
// @receiver e
// @param msg
func (e *raftElection) log(msg ...any) {

	var role string
	switch e.role {
	case LeaderRole:
		role = "领导者"
	case CandidateRole:
		role = "候选者"
	case FollowerRole:
		role = "跟谁者"
	}
	log.Println(e.address, role, msg)
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

	for true {

		for e.role == LeaderRole {

			select {
			case signal := <-e.eleSignalChan:
				switch signal.SignalType {
				case BeLeaderSignal:
					e.log("be leader")

				// 有时leader也会受到投票请求
				case AgreeVoteSignal:
					e.role = FollowerRole
					e.heartbeat.Stop()

				case SendFirstPingSignal:
					go e.raftPing()
					e.heartbeat.Reset(e.heartbeatMs * time.Millisecond)
					e.logSignalChan <- &Signal{EntriesCheckSignal, nil}

				case PingSignal:
					ping := signal.Data.(*RequestPing)

					if ping.Team > e.term.Load() {
						e.role = FollowerRole
						e.term.Store(ping.Team)
						e.log("leader ping, update Role")
						e.pingTime = ping.Time
						e.pongTime = time.Now().UnixMicro()
						e.logSignalChan <- &Signal{EntriesQuitSyncSignal, nil}
					}

				case QuitSignal:
					goto Quit
				}

			// start ping
			case <-e.heartbeat.C:
				go e.raftPing()
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
					e.log(time.Now().Format(time.StampMilli) + "start election")
					go e.raftRequestVote()

				case BadLeaderSignal:
					team := signal.Data.(int32)
					if e.term.Load() < team {
						e.term.Store(team)
					}

				case BeLeaderSignal:
					e.role = LeaderRole
					e.log(" CandidateRole BeLeader")
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
					e.log("re leader ping")
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
