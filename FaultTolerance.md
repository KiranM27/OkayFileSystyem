# Fault Tolerance 

1. Primary Chunk Server Fails (consistency)
- During append phase 
- During Commit phase  

2. Secondary Chunk Server Fails (consistency)
- During append phase 
- During Commit phase 

    Points 1 and 2: Since we follow a pipelined method for the transmission of messages, the client will not receive an ACK_COMMIT or an ACK_APPEND if any of the replica chunk servers fail during the operation. When the client doesn't receive an ACK at the end of a pre determined timeout, then the client assumes that the operation has failed and will try again later.

<hr>

3. Client tries to write to a file that is being replicated (consistency)

    The master doesn't respond to an append request unless and untill there are 3 replica servers corresponding to each chunk. Hence, if a client makes a write request to a chunk when the chunk is being replicated, the master will not respond and the client will timeout out and try again later.

    On the other hand, if the client has already recieved the list of 3 replica servers and if one of the replica servers fails during the write operation, then this falls under case 1 and 2 - which have already been accounted for.

<hr>

4. One of the chunk servers fails (availability)

    If one of the chunk server fails, it will be caught by the heartbeat mechanism that we have implemented. When a failed chunk server is detected, the master will begin the replication process and find a new chunk server for each of the chunks that went down with the chunk server. Once a chunk has been replicated in a new chunk server, the meta data of the master is updated so that new write requests to the chunk can be processed.

<hr>

5. One of the Chunk Server fails during a read (availability)

    Case 1: The client has already queried the master for a list of replica servers. A failure at this point would result in the client requesting the required data from the remianing chunk servers. If all three of the chunk servers have failed, then the client will query the master server again. If the master was able to replicate the chunks, then it will respond with the list of chunk servers, else we can assume that the chunk has been lost forever.

    Case 2: The client requests for the chunk servers after one of the chunk servers has failed. In this case, the master responds with the list of remaining chunk servers and tries to replicate the failed ones in the mean time.

<hr>