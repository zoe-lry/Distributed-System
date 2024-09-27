
## Run test
```
cd src/raft
go test
```

# Lab3 A
## Test 3A
```
cd src/raft
go test -run 3A
```

## Test Result
```
Test (3A): initial election ...
  ... Passed --   3.5  3   44   11366    0
Test (3A): election after network failure ...
  ... Passed --   8.1  3  136   29310    0
Test (3A): multiple elections ...
  ... Passed --   8.7  7  570  123626    0
PASS
ok      6.5840/raft     21.627s
```
## Reference
https://pdos.csail.mit.edu/6.824/labs/lab-raft.html

https://pdos.csail.mit.edu/6.824/papers/raft-extended.pdf



# Lab3A Raft Implementation
1. Fill in `RequestVoteArgs` and `RequestVoteReply` refer to Figure 2 in the paper
2. define `AppendEntryArgs` and `AppendEntryReply` refer to Figure 2
3. Fill in `Raft` struct and Modify `Make()` refer to Figure 2
4. Implement `Ticker()` to either call `SendHeartbeat()` or call `StartElection()`
    1. if the `heartbeat`timeout, call `SendHeartbeat()`
    2. if the `election` timeout, change the state to `candidate` and call `StartElection()`

5. Implement `SendHeartbeat()` method.
    1.  send `AppendEntry` to every peer except itself.

6. Implement `StartElection()`
    1. send `RequestVote` to every peer except itself.
    2. count the number of votes
    3. if the count is greater than half, change the state to `Learder`
    4. if the other `Follower` has higher term, change the `Candidate` state back to `Follower`

7. add `ChangeState` method
8. `AppendEntries`  Reply false if term < currentTerm 
9. `RequestVote` 
    1. Reply false if term < currentTerm (§5.1)
    2. If votedFor is null or candidateId, and  candidate’s log is at least as up-to-date as receiver’s log, grant vote (§5.2, §5.4)



