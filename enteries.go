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
	"sync/atomic"
	"time"
)

// raftEntries
// @Description: Entries Module
type raftEntries struct {
	*Node
	pEntries    []*Entries    // 模拟数据
	nextIndex   atomic.Uint32 // 数据下标? 待提交的下标
	matchIndex  atomic.Uint32 // 已经提交的下标
	commitIndex atomic.Uint32 // 已经提交的下标
	makeIndex   atomic.Int32

	lastLogIndex uint32
	lastLogTeam  uint32

	logSignalChan   chan *Signal
	entriesQuitChan NullChan
}

// NewRaftEntries
// @param node
// @return *raftEntries
func NewRaftEntries(node *Node) *raftEntries {
	e := new(raftEntries)
	e.Node = node
	e.logSignalChan = make(chan *Signal, SignalSize)
	e.entriesQuitChan = make(NullChan, 1)

	e.pEntries = make([]*Entries, 0)
	e.nextIndex.Store(0)
	e.matchIndex.Store(0)
	e.commitIndex.Store(0)

	go e.start()
	return e
}

// getSyncLogNodes
// @receiver e
// @param nextIndex
// @param matchIndex
// @return LogSyncNodes
func (e *raftEntries) getSyncLogNodes(nextIndex, matchIndex uint32) LogSyncNodes {

	checkNodes := make(LogSyncNodes)
	for _, node := range e.Nodes() {

		// 默认都给我投票了的,初始化和我一样. 检查发送的数据是判断这个来的. 这个nextIndex==1,PrevTerm=0,PrevIndex=0
		syncNode := NewSyncLogNode(node, nextIndex, matchIndex)
		syncNode.SetCommitIndex(e.commitIndex.Load())
		checkNodes[node.name] = syncNode
	}
	return checkNodes
}

// AddEntries
// @receiver e
// @param data
func (e *raftEntries) AddEntries(data any) {

	e.logSignalChan <- &Signal{EntriesAddSignal, data}
}

// _AddEntries
// @receiver e
// @param entries
func (e *raftEntries) _AddEntries(entries *Entries) bool {

	e.pEntries = append(e.pEntries, entries)
	e.lastLogIndex = e.nextIndex.Load()
	e.lastLogTeam = entries.Team
	e.nextIndex.Add(1)
	return true
}

func (e *raftEntries) _CommitEntries(cIndex uint32) {

	ics := e.commitIndex.Load()
	ice := ics + cIndex

	// perfect data
	if ics >= cIndex {
		return
	}

	//if maxIndex := e.nextIndex.Load(); ice > maxIndex {
	//	ice = maxIndex
	//}

	if maxIndex := uint32(len(e.pEntries)); ice > maxIndex {
		ice = maxIndex
	}

	// Commit Entries
	if e.name == ":9135" {
		fmt.Println("cimut ud = ", ics, ice, len(e.pEntries))
	}

	for _, entries := range e.pEntries[ics:ice] {
		entries.Commited = true
		entries.CommitN++
		e.commitIndex.Add(1)
	}

}

// newSyncRequestLog
// @receiver e
// @param rLog
// @return *SyncRequestLog
func (e *raftEntries) newSyncRequestLog(rLog *RequestLog) *SyncRequestLog {

	return &SyncRequestLog{
		RLog: rLog,
		Sync: make(chan bool),
	}
}

// checkEntriesResult
// @receiver e
// @param entries
func (e *raftElection) checkEntriesResult(entries *Entries) {

	if !entries.Commited && entries.CommitN > e.halfLen() {
		entries.Commited = true
		e.commitIndex.Add(1)
	}
}

func (e *raftElection) checkEntriesAll(list []*Entries) {

	for _, entries := range list {
		e.checkEntriesResult(entries)
	}
}

func (e *raftEntries) raftAppendEntries(logNode *LogSyncNode) {

	for true {

		select {

		case <-e.entriesQuitChan:
			goto Quit

		default:

			// 9135 matchindex = 0,  需要同步 下标0
			// 9137 matchindex = 1 , 需要同步 下标2
			// 9137 matchindex = 2 , 需要同步 下标3

			if !logNode.lastReq {
				<-time.After(e.heartbeatMs * time.Millisecond)
			}

			req := logNode.generateNextRpcRequestLog(e)

			if logNode.checkMatch {
				req.Log.Entries = e.getEntriesByIndex(logNode.nextIndex)

				if logNode.nextIndex == e.nextIndex.Load() {
					<-time.After(e.heartbeatMs * time.Millisecond)
				} else {
					logNode.nextIndexAdd(Uint32PlusOne)
				}

			} else {
				logNode.nextIndexAdd(Uint32SubOne)
			}

			e.logger.Debug("start sync log:", logNode.lastReq, req.Log)

			e.raftRpcServer.sendSyncSignal(req, logNode.SlaveNode, func(sNode *SlaveNode, err error, res *RpcResponse) {

				//fmt.Println(logNode.name+" error = ", err, res.Error)
				//fmt.Println("return:", fmt.Sprintf("%+v", res.Log))

				//logNode.lastReq = false
				//res.Log = nil

				if res.Log == nil {
					return
				}

				if res.Log.Success {
					logNode.matchIndex = res.Log.MatchIndex

					if req.Log.Entries != nil {
						e.logSignalChan <- &Signal{EntriesCommitSignal, req.Log.Entries}
					}
				}

				/*
					// check entries
					if !res.Log.Success && !logNode.checkMatch {
						logNode.nextIndexAdd(Uint32SubOne)
					}

					// sync entries
					if res.Log.Success && logNode.checkMatch {
						logNode.nextIndexAdd(Uint32PlusOne)
					}

				*/

				// start sync entries
				if res.Log.Success && !logNode.checkMatch {
					logNode.checkMatch = true

					// sync next nextIndex
					if res.Log.MatchIndex != 0 {
						logNode.nextIndex++
					}

					// sync 0 nextIndex
					if res.Log.MatchIndex == 0 {
						logNode.nextIndex = 0
					}
				}
			})

		}
	}

Quit:
}

func (e *raftEntries) startAppendEntries() {

	e.checkEntriesAll(e.pEntries)

	e.logger.Debug(e.name, "leader nindex = ", e.nextIndex.Load(), "commitindex = ", e.commitIndex.Load(), "datalen = ", len(e.pEntries))

	logNodes := e.getSyncLogNodes(e.nextIndex.Load(), 0)
	for _, logNode := range logNodes {
		go e.raftAppendEntries(logNode)
	}

}

func (e *raftEntries) receiveLogCallBassck(log *RequestLog, resLog *ResponseLog) {

}

// receiveLogCallBack
// @receiver e
// @param log
// @param resLog
func (e *raftEntries) receiveLogCallBack(log *RequestLog, resLog *ResponseLog) {

	e.eTimeoutReset()
	//e.logSignalChan <- &Signal{EntriesReceiveSignal, e.newSyncRequestLog(log)}

	// check log
	if log.Entries != nil {
		//e.logger.Debug(e.address, "sync nindex = ", e.nextIndex.Load(), "lastLogIndex = ", e.lastLogIndex, "commitindex = ", e.commitIndex.Load(), "leader ci", log.CommitIndex)
	} else {
		//e.logger.Debug(e.address, "check nindex = ", e.nextIndex.Load(), "lastLogIndex = ", e.lastLogIndex, "commitindex = ", e.commitIndex.Load(), "leader ci", log.CommitIndex)
	}

	e.nextIndex.Store(log.NextIndex)
	prevTerm, prevIndex := e.getPrevEntriesByNextIndex(e.nextIndex.Load())
	//prevTerm, prevIndex := e.getPrevEntriesByNextIndex(log.NextIndex)

	// 如果在不同日志中的两个条目有着相同的索引和任期号，则它们所存储的命令是相同的。
	// 如果在不同日志中的两个条目有着相同的索引和任期号，则它们之间的所有条目都是完全一样的。
	if prevTerm == log.PrevTerm && prevIndex == log.PrevIndex {

		if log.Entries != nil {
			slog := e.newSyncRequestLog(log)
			e.logSignalChan <- &Signal{EntriesReceiveSignal, slog}
			resLog.Success = <-slog.Sync
			resLog.MatchIndex = log.PrevIndex + 1
		} else {
			resLog.Success = true
			resLog.MatchIndex = log.PrevIndex
		}

		if log.CommitIndex != 0 {
			e._CommitEntries(log.CommitIndex)
		}

		resLog.Team = e.term.Load()
		e.matchIndex.Store(resLog.MatchIndex)

		if resLog.Success {
			//e.nextIndex.Add(Uint32PlusOne)
		}

	} else {

		//e.nextIndex.Add(Uint32SubOne)
		resLog.Team = log.Team
		resLog.Success = false
		resLog.MatchIndex = 0
	}

	// not a voting node
	if e.term.Load() != log.Team {
		e.term.Store(log.Team)
	}
}

func (e *raftEntries) getEntriesByIndex(n uint32) *Entries {

	if n > e.lastLogIndex || n < 0 {
		return nil
	}

	if e.lastLogIndex == 0 && e.lastLogTeam == 0 {
		return nil
	}

	return e.pEntries[n]
}

func (e *raftEntries) getPrevEntriesByNextIndex(nextIndex uint32) (prevTerm, prevIndex uint32) {

	prevIndex = nextIndex - 1
	prevEntries := e.getEntriesByIndex(prevIndex)

	if prevEntries != nil {
		prevTerm = prevEntries.Team
	} else {
		prevIndex = 0
	}

	return
}

func (e *raftEntries) start() {

	if startSignal := <-e.logSignalChan; startSignal.SignalType != StartSignal {
		panic(ErrStartSignal)
	}

	for {
		select {
		case signal := <-e.logSignalChan:
			switch signal.SignalType {

			case EntriesCheckSignal:
				go e.startAppendEntries()

			case EntriesQuitSyncSignal:
				e.entriesQuitChan <- NullSignal

			// leader commit
			case EntriesCommitSignal:
				entries := signal.Data.(*Entries)
				entries.CommitN++
				e.checkEntriesResult(entries)

			// leader add
			case EntriesAddSignal:
				entries := &Entries{false, 1, e.term.Load(), signal.Data}
				e._AddEntries(entries)
				e.checkEntriesResult(entries)

			case EntriesReceiveSignal:
				log := signal.Data.(*SyncRequestLog)
				log.Sync <- e._AddEntries(log.RLog.Entries)

			}

		}
	}
}
