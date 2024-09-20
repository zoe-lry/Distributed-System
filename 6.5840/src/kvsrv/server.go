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


}


func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	// Your code here.
	kv.mu.Lock();
	defer kv.mu.Unlock();
	reply.Value = kv.data[args.Key];
}

func (kv *KVServer) Put(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	kv.mu.Lock();
	defer kv.mu.Unlock();
	kv.data[args.Key] = args.Value;
}

func (kv *KVServer) Append(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	kv.mu.Lock();
	defer kv.mu.Unlock();
	oldValue := kv.data[args.Key];
	kv.data[args.Key] = oldValue + args.Value;
	reply.Value = oldValue;
}

func StartKVServer() *KVServer {
	kv := new(KVServer)

	// You may need initialization code here.
	kv.data = make(map[string]string);
	return kv
}
