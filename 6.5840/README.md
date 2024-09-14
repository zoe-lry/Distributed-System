# Usage
## Run test
- go to main
  ```
  cd src/main
  ```
- Build the plugin
    ```
    go build -buildmode=plugin ../mrapps/wc.go
    ```
- run test
  ```angular2html
  bash test-mr.sh 
    ```
Output of the test sould be:
  ```
    zoeli@zoedeMacBook-Pro main % go build -buildmode=plugin ../mrapps/wc.go
    zoeli@zoedeMacBook-Pro main % bash test-mr.sh                           
    *** Cannot find timeout command; proceeding without timeouts.
    *** Starting wc test.
    --- wc test: PASS
    *** Starting indexer test.
    --- indexer test: PASS
    *** Starting map parallelism test.
    --- map parallelism test: PASS
    *** Starting reduce parallelism test.
    --- reduce parallelism test: PASS
    *** Starting job count test.
    --- job count test: PASS
    *** Starting early exit test.
    --- early exit test: PASS
    *** Starting crash test.
    --- crash test: PASS
    *** PASSED ALL TESTS
  ```

## Run simulating the coordinator and workers
- go to main
  ```
  cd src/main
  ```
- Build the plugin
    ```
    go build -buildmode=plugin ../mrapps/wc.go
    ```
- run `mrcoordinator.go`;
    ```angular2html
    go run mrcoordinator.go pg-*.txt
    ```
- open different terminals and run multiple `mrworker.go`; 
    ```angular2html
    cd src/main
    go run mrworker.go wc.so
    ```

## Reference
- 《Distributed Systems》(6.824)LAB1(mapreduce)                 
原文链接：https://blog.csdn.net/weixin_44520881/article/details/109641515
https://blog.csdn.net/weixin_44520881/article/details/109641515?spm=1001.2014.3001.5501

- 
# 1. Introduction to Implementing own MapReduce

## 1.1 Try Sample use

> 1. build the `wc.go` as a plugin. After built, the plugin name will be `wc.so`.
> 2. remove the existing outputs
> 3. run `mrsequential.go` with plugin `wc.so` and files `pg-*.txt`.
> 4. open output name `mr-out-0`
>
> the output would be the word with its frequency in those files.

```angular2html
$ cd ~/6.5840
$ cd src/main
$ go build -buildmode=plugin ../mrapps/wc.go
$ rm mr-out*
$ go run mrsequential.go wc.so pg*.txt
$ more mr-out-0
A 509
ABOUT 2
ACT 8
...
```

## 1.2 main tasks

> Your job is to implement a distributed MapReduce, consisting of two programs,
> `the coordinator` and `the worker`.
> There will be just one coordinator process, and one or more worker processes executing in parallel.
> In a real system the workers would run on a bunch of different machines, but for this lab you'll run them all on a single machine.
>
> `The workers` will talk to the coordinator via RPC.
> Each worker process will, in a loop, ask the coordinator for a task,
> read the task's input from one or more files, execute the task,
> write the task's output to one or more files, and again ask the coordinator for a new task.
>
> `The coordinator` should notice if a worker hasn't completed its task in a reasonable amount of time
> (for this lab, use ten seconds), and give the same task to a different worker.

> We have given you a little code to start you off.
> The "main" routines for `the coordinator` and `worker` are in `main/mrcoordinator.go` and `main/mrworker.go`;

> **don't change these files**. You should put your implementation in `mr/coordinator.go`, `mr/worker.go`, and `mr/rpc.go`.

## 2.2 A few rules

A few rules:

- The map phase should divide the intermediate keys into buckets for `nReduce` reduce tasks,
  where `nReduce` is the number of reduce tasks
  -- the argument that `main/mrcoordinator.go` passes to `MakeCoordinator()`.
  Each mapper should create `nReduce` intermediate files for consumption by the reduce tasks.

- The worker implementation should put the output of the X'th reduce task in the file `mr-out-X`.

- A `mr-out-X` file should contain one line per Reduce function output.
  The line should be generated with the Go `%v %v` format, called with the key and value.
  Have a look in `main/mrsequential.go` for the line commented "this is the correct format".
  The test script will fail if your implementation deviates too much from this format.

- You can modify `mr/worker.go`, `mr/coordinator.go`, and `mr/rpc.go`.
  You can temporarily modify other files for testing, but make sure your code works with the original versions;
  we'll test with the original versions.

- The worker should put intermediate Map output in files in the current directory, where your worker can later read them as input to Reduce tasks.

- `main/mrcoordinator.go` expects `mr/coordinator.go` to implement a `Done()` method
  that returns true when the MapReduce job is completely finished;
  at that point, `mrcoordinator.go` will exit.
  When the job is completely finished, the worker processes should exit.
  A simple way to implement this is to use the return value from `call()`:
  if the worker fails to contact the coordinator,
  it can assume that the coordinator has exited because the job is done, so the worker can terminate too.
  Depending on your design, you might also find it helpful to have a "please exit" pseudo-task that the coordinator can give to workers.

## 1.3 Hints

- The Guidance page has some tips on developing and debugging.

  Then modify the coordinator to respond with the file name of an as-yet-unstarted map task.
  Then modify the worker to read that file and call the application Map function, as in `mrsequential.go`.

- The application Map and Reduce functions are loaded at run-time using the Go plugin package, from files whose names end in .so.
  If you change anything in the mr/ directory, you will probably have to re-build any MapReduce plugins you use, with something like go build -buildmode=plugin ../mrapps/wc.go
  This lab relies on the workers sharing a file system. That's straightforward when all workers run on the same machine, but would require a global filesystem like GFS if the workers ran on different machines.
- A reasonable naming convention for intermediate files is mr-X-Y, where X is the Map task number, and Y is the reduce task number.
  The worker's map task code will need a way to store intermediate key/value pairs in files in a way that can be correctly read back during reduce tasks. One possibility is to use Go's encoding/json package. To write key/value pairs in JSON format to an open file:

```
    enc := json.NewEncoder(file)
    for _, kv := ... {
        err := enc.Encode(&kv)
```

and to read such a file back:

```
    dec := json.NewDecoder(file)
    for {
        var kv KeyValue
        if err := dec.Decode(&kv); err != nil {
            break
        }
        kva = append(kva, kv)
    }
```

The map part of your worker can use the ihash(key) function (in worker.go) to pick the reduce task for a given key.
You can steal some code from `mrsequential.go` for reading Map input files, for sorting intermedate key/value pairs between the Map and Reduce, and for storing Reduce output in files.
The coordinator, as an RPC server, will be concurrent; don't forget to lock shared data.
Use Go's race detector, with go run -race. test-mr.sh has a comment at the start that tells you how to run it with -race. When we grade your labs, we will not use the race detector. Nevertheless, if your code has races, there's a good chance it will fail when we test it even without the race detector.
Workers will sometimes need to wait, e.g. reduces can't start until the last map has finished. One possibility is for workers to periodically ask the coordinator for work, sleeping with time.Sleep() between each request. Another possibility is for the relevant RPC handler in the coordinator to have a loop that waits, either with time.Sleep() or sync.Cond. Go runs the handler for each RPC in its own thread, so the fact that one handler is waiting needn't prevent the coordinator from processing other RPCs.
The coordinator can't reliably distinguish between crashed workers, workers that are alive but have stalled for some reason, and workers that are executing but too slowly to be useful. The best you can do is have the coordinator wait for some amount of time, and then give up and re-issue the task to a different worker. For this lab, have the coordinator wait for ten seconds; after that the coordinator should assume the worker has died (of course, it might not have).
If you choose to implement Backup Tasks (Section 3.6), note that we test that your code doesn't schedule extraneous tasks when workers execute tasks without crashing. Backup tasks should only be scheduled after some relatively long period of time (e.g., 10s).
To test crash recovery, you can use the mrapps/crash.go application plugin. It randomly exits in the Map and Reduce functions.
To ensure that nobody observes partially written files in the presence of crashes, the MapReduce paper mentions the trick of using a temporary file and atomically renaming it once it is completely written. You can use ioutil.TempFile (or os.CreateTemp if you are running Go 1.17 or later) to create a temporary file and os.Rename to atomically rename it.
test-mr.sh runs all its processes in the sub-directory mr-tmp, so if something goes wrong and you want to look at intermediate or output files, look there. Feel free to temporarily modify test-mr.sh to exit after the failing test, so the script does not continue testing (and overwrite the output files).
test-mr-many.sh runs test-mr.sh many times in a row, which you may want to do in order to spot low-probability bugs. It takes as an argument the number of times to run the tests. You should not run several test-mr.sh instances in parallel because the coordinator will reuse the same socket, causing conflicts.
Go RPC sends only struct fields whose names start with capital letters. Sub-structures must also have capitalized field names.
When calling the RPC call() function, the reply struct should contain all default values. RPC calls should look like this:

```
reply := SomeType{}
call(..., &reply)
```

without setting any fields of reply before the call. If you pass reply structures that have non-default fields, the RPC system may silently return incorrect values.

## 1.4 思路整理

通读一下`mrsequential.go`和`wc.go`两个文件：
`wc.go`包含一个`Map`和一个`Reduce`方法

- `Map`: return value is a slice of key/value pairs. 返回一个类似 counter 的东西
- `Reduce`: 返回某个词出现的次数

`mrsequential.go` 主要是读取每个文件，把每个文件的切片都追加到`intermediate`里面，在用`sort`排序

## 1.5 Initial Run

- Modify the `mr/worker.go`'s `Worker()` to send an RPC to the coordinator asking for an task:
  - uncomment `CallExample()` to send the Example RPC to the coordinator.
- Build the plugin
    ```
    go build -buildmode=plugin ../mrapps/wc.go
    ```
- run `mrcoordinator.go`;
- ```angular2html
    go run mrcoordinator.go pg-*.txt
    ```
 - run `mrworker.go`; 
 - ```angular2html
        go run mrworker.go wc.so
启动`mrcoordinator.go`，就已经通过`MakeCoordinator`启动服务器，监听外来请求；
`CallExample()`会向已经开启的监听套接字发出请求，建立链接；
`c, err := rpc.DialHTTP("unix", sockname)` 得到的c是一个RPC客户端对象，
他有个方法`c.all()`可以调用远程的方法，即`Example`，作出应答，讲 arg + 1 传给 reply，并打印出来，说明链接成功。

```angular2svg
zoeli@zoedeMacBook-Pro main % go run mrworker.go wc.so
reply.Y 100
```

# 2. Implementation
## 2.1 Task 1 -- Modify `Worker()` to send an RPC to Coordinatior
One way to get started is to modify `mr/worker.go`'s `Worker()` to send an RPC to the coordinator asking for a task.
-   Create a `Task` and a `Reply` struct in `rpc.go`
## 2.2 Add WriteMapOutput method 
to distribute the keyvalue pair to each reduce task and write into files named `rm-X-Y` where X is the Map task number, and Y is the reduce task number.
Y = hash(key) % nReduce. 

Hint: The map part of your worker can use the `ihash(key)` function (in `worker.go`) to pick the reduce task for a given key.

## 2.3 Adding reponse from Worker to Coordinator.
While the Worker finished one task, call `FinishTask` in the coordinator to inform it. and the coordinate would update the overall status and check if the current stage has been completes.

e.g. 
if all map tasks are completeded then switch to reduce tasks; 

or if all reduce tasks have completed. then close the Coordinator

## 2.4 Adding Reduce haldler in Worker.go
After the Map tasks complete, start the reduce tasks.

Coordinator assigns reduce task to worker (use `status` to identify if a given task is map or reduce)

worker receive a reduce task and CallReduce method.
it will read the file name `out-X-Y` where Y is assigned by `Coordinator` and X is the loop from 0 to NMap (the number of files).

## 2.5 Adding Machine Number distribute
Keep tracking the number of Worker

## 2.6 Adding TimeTick()
to monitor if the machines are running properly. If it runout of time (10 seconds in this case). return the status to 0 and reassign the task to the next machine.
