package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	for {
		// sleep one second before calling each task
		time.Sleep(time.Second)
		request := TaskRequest{} // declare an argument structure.
		reply := TaskResponse{} // declare a reply structure.

		// send the RPC request, wait for the reply.
		// the "Coordinator.GetTask" tells the
		// receiving server that we'd like to call
		// the GetTask() method of struct Coordinator.
		CallGetTask(&request, &reply)

		if reply.Status == 0 { 	// if the status is map
			// Get the filename by calling the CallGetTask
			kva := CallMap(mapf, reply.FileName, reply.NReduce)
			// split the Map into nReduce tasks
			ok := WriteMapOutput(kva, reply.TaskNumber, reply.NReduce)
			if ok {
				// reply.Status = 2
			}
		} else if (reply.Status ==  1) {

		} else if (reply.Status == 2) {

		}
		// pass the task status and task number to corrdinator 
		finishReply := FinishReply{Status: reply.Status, TaskNumber: reply.TaskNumber }
		finishResponse := FinishResponse{}
		// call FinishTask in coordinator
		CallFinishTask(&finishReply, &finishResponse)
		// get the overall status. if all tasks are done. break the loop
		// otherwise call another task
		if finishReply.Status == 2 {
			break
		}

	}
	

	// Your worker implementation here.

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()

}



// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// // the RPC argument and reply types are defined in rpc.go.
func CallGetTask(args *TaskRequest, reply *TaskResponse)  {
	ok := call("Coordinator.GetTask", &args, &reply)
	if ok {
		fmt.Printf("reply.Name %s\n", reply.FileName)
	} else {
		fmt.Printf("call failed!\n")
	}
	
}

func CallFinishTask(reply *FinishReply, response *FinishResponse)  {
	ok := call("Coordinator.FinishTask", &reply, &response)
	if ok {
		fmt.Printf("finish task call --- status: %v\n", reply.Status)
	} else {
		fmt.Printf("call failed!\n")
	}
	
}
// read the file and turn into keyValue slice
func CallMap(mapf func(string, string) []KeyValue, filename string, nReduce int) []KeyValue{
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open %v", filename)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", filename)
	}
	file.Close()
    // Call the map function
    kva := mapf(filename, string(content))

    // Return the kva slice if needed
    return kva
}

// Create a 2D slice (slice of slices) to store key-value pairs for each reduce task
// for eacg keyValue, ihash number % nReduce
// 
func WriteMapOutput(kva []KeyValue, taskNumber int, nReduce int) bool {
	buf := make ([][] KeyValue, nReduce)
	for _, kv := range(kva){
		bucket := ihash(kv.Key) % nReduce
		buf[bucket] = append(buf[bucket], kv)
	}
	for no, kv_pairs := range(buf) {
		mapOutputFilename := "mr-" + strconv.Itoa(taskNumber) + "-" + strconv.Itoa(no)
		// Create a temp file named mr-map-random name
		// Once the data is written, the temporary file is renamed to mr-X-Y to store the partitioned data for that reduce task.
		// To ensure that nobody observes partially written files in the presence of crashes
		tmpMapOutFile, error := os.CreateTemp("", "mr-map-*")
		if error != nil {
			log.Fatalf("Fail to open tmpMapOutFile.")
		}
		// encode the kv_pair to the temp file
		enc := json.NewEncoder(tmpMapOutFile)
		err := enc.Encode(kv_pairs)
		if err != nil {
			log.Fatalf("Fail to encode tmpMapOutFile.")
			return false
		}
		tmpMapOutFile.Close()
		os.Rename(tmpMapOutFile.Name(), mapOutputFilename)
	}
	return true
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")

	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)

	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
