package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strconv"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

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
	machineId := 0
	for {
		// sleep one second before calling each task
		time.Sleep(time.Second)
		request := TaskRequest{MachineId: machineId} // declare an argument structure.
		response := TaskResponse{} // declare a reply structure.

		// send the RPC request, wait for the reply.
		// the "Coordinator.GetTask" tells the
		// receiving server that we'd like to call
		// the GetTask() method of struct Coordinator.
		CallGetTask(&request, &response)
		machineId = response.MachineId

		if response.Status == 0 { 	//process Map
			// Get the filename by calling the CallGetTask
			kva := CallMap(mapf, response.FileName, response.NReduce)
			// split the Map into nReduce tasks
			ok := WriteMapOutput(kva, response.TaskNumber, response.NReduce)
			if ok {
				// TODO  if the task didnt finish properly, reply to the coordinator?
				// reply.Status = 2 
			}
		} else if (response.Status ==  1) { // process reduce
			// Call reduce plugin to complete 
			 CallReduce(reducef, response.TaskNumber, response.NMap)


		} else if (response.Status == 2) {

		}
		
		// pass the task status and task number to corrdinator 
		finishRquest := FinishReply{Status: response.Status, TaskNumber: response.TaskNumber }
		finishResponse := FinishResponse{}
		// call FinishTask in coordinator
		CallFinishTask(&finishRquest, &finishResponse)
		// get the overall status. if all tasks are done. break the loop
		// otherwise call another task
		if finishRquest.Status == 2 {
			break
		}
	}
	
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

// the RPC argument and reply types are defined in rpc.go.
func CallGetTask(request *TaskRequest, response *TaskResponse)  {
	ok := call("Coordinator.GetTask", &request, &response)
	if ok {
		fmt.Printf("GET TASK ---Machine ID: %v,  task type %v, filename %s\n", response.MachineId, response.Status, response.FileName)
	} else {
		fmt.Printf("call failed!\n")
	}
	
}

// Task finished and feed back to coordinator
func CallFinishTask(reply *FinishReply, response *FinishResponse)  {
	ok := call("Coordinator.FinishTask", &reply, &response)
	if ok {
		fmt.Printf("FINISHED --- task type %v, TaskNumber %v\n", reply.Status, reply.TaskNumber)
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
			log.Fatalf("Fail to open tmpMapOutFile: %v", error)
		}
		// encode the kv_pair to the temp file
		enc := json.NewEncoder(tmpMapOutFile)
		err := enc.Encode(kv_pairs)
		if err != nil {
			log.Fatalf("Fail to encode tmpMapOutFile: %v", err)
			return false
		}
		tmpMapOutFile.Close()
		os.Rename(tmpMapOutFile.Name(), mapOutputFilename)
	}
	return true
}

func CallReduce(reducef func(string, []string) string,taskNumber int , nMap int) {
	intermediate := []KeyValue{}
	// read each file mr-mapnumber-tasknumber and reduce
	for i := 0; i < nMap; i++ {
		reduceFilename := "mr-" + strconv.Itoa(i) + "-" + strconv.Itoa(taskNumber)
		file, err := os.OpenFile(reduceFilename, os.O_RDONLY, 0777)
		if err != nil {
			log.Fatalf("cannot open reduce task file name: %v", reduceFilename)
		}
		dec := json.NewDecoder(file)
		for{
			var kv []KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			intermediate = append(intermediate, kv...)
		}
		file.Close()
	}

	// Sort the key value pair
	sort.Sort(ByKey(intermediate))
	outFilename := "mr-out-" + strconv.Itoa(taskNumber)
	tmpOutFilename, _ := os.Create(outFilename)

	// call Reduce on each distinct key in intermediate[],
	// and print the result to mr-out-0.
	i := 0
	for i < len(intermediate) {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		output := reducef(intermediate[i].Key, values)

		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(tmpOutFilename, "%v %v\n", intermediate[i].Key, output)
		i = j
	}

	tmpOutFilename.Close()
	os.Rename(tmpOutFilename.Name(), outFilename) // Rename the file 

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
