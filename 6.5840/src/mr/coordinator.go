package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Task struct {
	FileName string
	Status int // 0-start 1-running 2-end
	Runtime int // Time of running
	MachineId int 
}

type Coordinator struct {
	// Your definitions here.
	Status int 	// track the task type 0-Map & 1-Reduce 2-Done
	MapTasks map[int]*Task
	ReduceTasks map[int]*Task
	MachineNum int
	NMap int	// 最大并行map的个数， 哈希的个数
	NReduce int // 最大并行的reduce的个数，哈希的个数
	Mu sync.Mutex //只能有一个worker访问
}

// Monitor the running time of each task.
// If the task running time exceed 10 seconds. 
// set the status back to 0 and reassign to other Worker the next time
func (c *Coordinator) TimeTick() {
	c.Mu.Lock()
	if (c.Status == 0) {
		for _, task := range(c.MapTasks) {
			if task.Status == 1 {
				task.Runtime ++
				if (task.Runtime >= 10) {
					task.Status = 0
					task.Runtime = 0
				}
			}
		}
	} else if (c.Status == 1) {
		for _, task := range(c.ReduceTasks) {
			if task.Status == 1 {
				task.Runtime ++
				if (task.Runtime >= 10) {
					task.Status = 0
					task.Runtime = 0
				}
			}
		}
	}
	c.Mu.Unlock()
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) GetTask(request *TaskRequest, response *TaskResponse) error {
	c.Mu.Lock()
	response.Status = 2 
	response.NReduce = c.NReduce
	response.NMap = c.NMap
	// Assign a machine ID if not exist
	if (request.MachineId == 0) {
		c.MachineNum ++
		response.MachineId = c.MachineNum
	} else {
		response.MachineId = request.MachineId
	}
	// fmt.Printf("c.Status: %v, Machine Number: % v \n", c.Status, response.MachineId)
	if c.Status == 0 {
		for taskNumber, task := range(c.MapTasks) {
			if (task.Status == 0){
				task.Status = 1 	// start Running
				response.Status = 0 // 0 - Map task
				response.FileName = task.FileName
				response.TaskNumber = taskNumber
				break
			}
		}
	} else if c.Status == 1 {
		for taskNumber, task := range(c.ReduceTasks) {
			if (task.Status == 0){
				task.Status = 1 	// start Running

				response.TaskNumber = taskNumber
				response.Status = 1 // 1- Reduce task
				break
			}
		}
	} else {

	}

	c.Mu.Unlock()
	return nil
}

//  worker finished one task
func (c *Coordinator) FinishTask(reply *FinishRequest, response *FinishResponse) error {
	c.Mu.Lock()
	response.Status = 1
	if (reply.Status == 0) { // map
		c.MapTasks[reply.TaskNumber].Status = 2
		c.UpdateStatus()
	} else if (reply.Status == 1) { // reduce
		c.ReduceTasks[reply.TaskNumber].Status = 2
		c.UpdateStatus()
		if c.Status == 2 {
			response.Status = 2
		}
	}

	c.Mu.Unlock()
	return nil
}

// Check Status. switch tasks status
func (c *Coordinator) UpdateStatus() {
	if (c.Status == 0) {
		for _, task := range(c.MapTasks) {
			if (task.Status == 0 || task.Status == 1){
				c.Status = 0
				return
			}
		}
		c.Status = 1
		// fmt.Printf("TASKS CHANGE HERE c.Status v%: \n", c.Status)
		return
	} else if (c.Status == 1) {
		for _, task := range(c.ReduceTasks) {
			if (task.Status == 0 || task.Status == 1){
				c.Status = 1
				return
			}
		}
		c.Status = 2
		return
	} 
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
	c.TimeTick()

	// Your code here.
	// if the state == 2 means all tasks done
	if (c.Status == 2) {
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
	time.Sleep(time.Second)
	c := Coordinator{
		Status : 0,
		MapTasks: make(map[int]*Task),
		ReduceTasks: make(map[int]*Task),
		MachineNum: 0,
		NMap: len(files),
		NReduce: nReduce, // 最大并行的reduce的个数，哈希的个数
	}
	// put the file names into MapTasks
	for i, filename:= range files {
		c.MapTasks[i] = &Task{FileName: filename, Status: 0, Runtime: 0, MachineId: 0}
	}
	// create nReduce number of reduceTasks
	for i:=0; i < nReduce; i++ {
		c.ReduceTasks[i] = &Task{FileName: "", Status: 0, Runtime: 0, MachineId: 0}
	}

	c.server()
	return &c
}
