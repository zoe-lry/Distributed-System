package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.
type TaskRequest struct {
	MachineId int

}

type TaskResponse struct {
	FileName string
	TaskNumber int
	Status int // 0-map 1-reduce 2 no task
	NReduce int
	NMap int
	MachineId int
}

// when worker finished one task, send a reply to coordinator
type FinishReply struct {
	TaskNumber int
	Status int
}

// Coordination reply the status to worker.

type FinishResponse struct {
	Status int // 1-processing 2-finished
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
