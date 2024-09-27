package raft

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

// Debugging
const Debug = false

// log entry struct
type LogEntry struct {
	Term int
	Index int
	Command interface{}
}

// generate random election timeout
const ElectionTimeout = 1000
const heartbeaTimeout = 125

type LockedRand struct {
	mu   sync.Mutex
	rand *rand.Rand
}

func (r *LockedRand) Intn(n int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Intn(n)
}

var GlobalRand = &LockedRand{
	rand: rand.New(rand.NewSource(time.Now().UnixNano())),
}
// election timeout
func RandomElectionTimeout() time.Duration {
	return time.Duration(ElectionTimeout + GlobalRand.Intn(ElectionTimeout)) * time.Millisecond
}
// heartbeat timeout
func RandomHeartbeatTimeout() time.Duration {
	return time.Duration(heartbeaTimeout) * time.Millisecond
}

// Server State
// follower - 0; candidate - 1; leader - 2
type ServerState uint8
const (
	Follower ServerState = iota
	Candidate
	Leader
)

// Get the last log
func (rf *Raft) GetLastLog() LogEntry{
	return rf.logs[len(rf.logs)-1]
}

func DPrintf(format string, a ...interface{}) {
	if Debug {
		log.Printf(format, a...)
	}
}
