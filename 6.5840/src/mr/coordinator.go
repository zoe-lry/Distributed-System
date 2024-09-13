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
	fileName string
}

type Coordinator struct {
	// Your definitions here.
	tasks []Task
	i int
	size int

}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) GetTask(args *TaskRequest, reply *TaskResponse) error {
	if (c.Done()) {
		return fmt.Errorf("no more tasks available, all tasks are completed")
	}
	// Otherwise, assign the next available task
	reply.FileName = c.tasks[c.i].fileName
	// Move to the next task
	c.i ++
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
	if (c.size == c.i) {
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
		tasks: []Task{},
		i:		0,
		size: len(files)}
	
	for _, filename:= range files {
		task := Task{fileName: filename}
		c.tasks = append(c.tasks, task)
	}

	c.server()
	return &c
}
