package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

type Task struct {
	FileName string
	State int // 0-start 1-running 2-end
	Runtime int // Time of running
	MachineId int 

}

type Coordinator struct {
	// Your definitions here.
	State int 	// track the task type 0-Map & 1-Reduce 2-Done
	MapTasks map[int]*Task
	ReduceTasks map[int]*Task
	MachineNum int
	NMap int	// 最大并行map的个数， 哈希的个数
	NReduce int // 最大并行的reduce的个数，哈希的个数

}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) GetTask(args *TaskRequest, reply *TaskResponse) error {
	if (c.Done()) {
		return fmt.Errorf("no more tasks available, all tasks are completed")
	}
	// if state == 0, assign map task
	if c.State == 0 {
		for taskNumber, task := range(c.MapTasks) {
			if (task.State == 0){
				reply.FileName = task.FileName
				reply.TaskNumber = taskNumber
				break
			}
		}
	}
	// fmt.Printf("reply.Name %s\n", reply.Name)
	return nil
}

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 100
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	ret := false
	// Your code here.
	// if the state == 2 means all tasks done
	if (c.State == 2) {
		ret = true
	}
	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	// Your code here.
	c := Coordinator{
		State : 0,
		MapTasks: make(map[int]*Task),
		ReduceTasks: make(map[int]*Task),
		MachineNum: 0,
		NMap: len(files),
		NReduce: nReduce, // 最大并行的reduce的个数，哈希的个数
	}
	// put the file names into MapTasks
	for i, filename:= range files {
		c.MapTasks[i] = &Task{FileName: filename, State: 0, Runtime: 0, MachineId: 0}
	}
	// create nReduce number of reduceTasks
	for i:=0; i < nReduce; i++ {
		c.ReduceTasks[i] = &Task{FileName: "", State: 0, Runtime: 0, MachineId: 0}
	}

	c.server()
	return &c
}
