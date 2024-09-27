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


// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (3A, 3B).
	Term int			// Candidate's term
	CandidateId int 	// Candidate requesting vote
	LastLogIndex int 	// index of candidate’s last log entry
	LastLogTerm int		// term of candidate’s last log entry
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (3A).
	Term int		// currentTerm, for candidate to update itself
	VoteGranded bool	// true means candidate received vote
}
