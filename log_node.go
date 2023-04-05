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

// LogSyncNode
// @Description: sync entries.
type LogSyncNode struct {
	checkMatch            bool
	nextIndex, matchIndex uint32
	*SlaveNode
	*RequestLog
}

type LogSyncNodes = map[string]*LogSyncNode

// NewSyncLogNode
// @param sNode
// @param nextIndex
// @param matchIndex
// @return *LogSyncNode
func NewSyncLogNode(sNode *SlaveNode, nextIndex, matchIndex uint32) *LogSyncNode {
	n := new(LogSyncNode)

	n.nextIndex = nextIndex
	n.matchIndex = matchIndex

	// slice[waitSyncStart:waitSyncEnd]
	//n.waitSyncStart = 0
	//n.waitSyncEnd = nextIndex

	n.SlaveNode = sNode
	n.RequestLog = new(RequestLog)

	return n
}

// SetCommitIndex
// @receiver n
// @param index
func (n *LogSyncNode) SetCommitIndex(index uint32) {

	n.CommitIndex = index
}

// SetRequestLog
// @receiver n
// @param reqLog
func (n *LogSyncNode) SetRequestLog(reqLog *RequestLog) {

	n.RequestLog = reqLog
}

// SetMatchIndex
// @receiver n
// @param index
func (n *LogSyncNode) SetMatchIndex(index uint32) {

	n.matchIndex = index
}

// nextIndexAdd
// @receiver n
// @param num
// @return bool
func (n *LogSyncNode) nextIndexAdd(num uint32) bool {

	if !n.lastReq {
		return false
	}

	n.nextIndex += num
	return true
}

// generateNextRpcRequestLog
// @receiver n
// @param e
// @return *RpcRequest
func (n *LogSyncNode) generateNextRpcRequestLog(e *raftEntries) *RpcRequest {

	prevTerm, prevIndex := e.getPrevEntriesByNextIndex(n.nextIndex)
	n.RequestLog.PrevTerm = prevTerm
	n.RequestLog.PrevIndex = prevIndex
	n.RequestLog.Entries = nil
	n.RequestLog.Team = e.term.Load()
	n.RequestLog.CommitIndex = e.commitIndex.Load()
	n.RequestLog.NextIndex = n.nextIndex
	return &RpcRequest{Log: n.RequestLog}
}
