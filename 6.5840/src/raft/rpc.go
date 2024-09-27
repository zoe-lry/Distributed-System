package raft

type AppendEntryArgs struct{
	Term int			// leader’s term
	LeaderId int
	PrevLogIndex int	// index of log entry immediately preceding new ones
	PrevLogTerm int		// term of prevLogIndex entry
	Entries	[]LogEntry	// log entries to store (empty for heartbeat; may send more than one for efficiency)
	LeaderCommit int 	// leader’s commitIndex
}

type AppendEntryReply struct {
	Term int	// currentTerm, for leader to update itself
	Success bool // true if follower contained entry matching prevLogIndex and prevLogTerm
}