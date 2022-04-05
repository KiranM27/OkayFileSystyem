Every thing in MetaData must be a sync map:

- Upon appendRequest, we spin up a go routine to handle this and make changes to meta data. But if there is multiple clients asking to append then we will have a problem.

- Same thing with replicate. Each time we replicate, we spin up a go routine for the dead chunk server. If there is one chunk server then there wont be an issue.
  All ports still seem to be hard coded

```
func choose_3_random_chunkServers() []int {

	chunkServerArray := map[int]bool{
		8081: false,
		8082: false,
		8083: false,
		8084: false,
		8085: false,
	}
```

Fault Tolerance

1. Primary/Secondary we let heartbeat find out
2. During a write, one server fails
   - the client will continue to timeout as it can only successfully write when the primary sends it a commit
   - Secondary fails, primary will never send ACK_DATA_COMMIT to client
   - Primary fail, same thing as above
3. During a read, server fails
   - client timesout and reads from the next repli4ca
   - if no more to read from, request from master again
4. Client tires to write to chunk being replicated.
   - Check that length of Chunk server ids == 3 (or rather the portNo in the message ==4), else we make the client timeout. By that time we should already have done replication/are doing replication.

TODOS:

1. look into exec() and clean up the code. Right now, there are too many global variables used by multiple go routines which is not correct. We need to run each client/chunk/master on a NEW terminal and they need to manage their own ACKs.  This means we also need master to be able to spin up other chunk servers in new terminals via go code.
2. Source chunk server failing: need to ask the next node, but refer to the 1st point about having each chunkserver manage their own map instead of global map
3. Target CS fails: Master needs to take all the chunks it was replicating + all its own chunks and replicate it accordingly. 
4. Because we create a new ACKMap system for the client, we did not manage to update it for the chunking portion of the client. We had assumed a client handles writing 1 record at a time but this is not the case when he writes a file larger than a chunk size. But refer to point 1 again before diving into this. We need to manage the data management side first. 