package kvsrv

import (
	"log"
	"sync"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}


type KVServer struct {
	mu sync.Mutex
	
	// Your definitions here.
	data map[string]string; // key - value map
	record sync.Map

}


func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	// Your code here.
	kv.mu.Lock();
	defer kv.mu.Unlock();
	reply.Value = kv.data[args.Key];
}

func (kv *KVServer) Put(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	// hanld different message type
	// if type is 1, delete the record
	if args.MessageType == 1 {
		kv.record.Delete(args.Id)
		return
	}
	// if the message has been handled, return the recorded result
	res, exist := kv.record.Load(args.Id)
	if exist{
		reply.Value = res.(string);
		return
	}
	// if not yet handled
	kv.mu.Lock();
	defer kv.mu.Unlock();
	kv.data[args.Key] = args.Value;
	kv.record.Store(args.Id, args.Value) // record the handled message
}

func (kv *KVServer) Append(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
		// hanld different message type
	// if type is 1, delete the record
	if args.MessageType == 1 {
		kv.record.Delete(args.Id)
		return
	}
	// if the message has been handled, return the recorded result
	res, exist := kv.record.Load(args.Id)
	if exist{
		reply.Value = res.(string);
		return
	}
	// if not yet handled
	kv.mu.Lock();
	defer kv.mu.Unlock();
	oldValue := kv.data[args.Key];
	kv.data[args.Key] = oldValue + args.Value;
	reply.Value = oldValue;
	kv.record.Store(args.Id, oldValue) // record the handled message
}

func StartKVServer() *KVServer {
	kv := new(KVServer)

	// You may need initialization code here.
	kv.data = make(map[string]string);
	kv.record = sync.Map{};
	return kv
}
