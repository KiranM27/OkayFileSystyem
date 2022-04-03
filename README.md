# 50.041 - Distributed Systems & Computing 
## Okay File System

A simplified version of the [Google File System (GFS)](https://static.googleusercontent.com/media/research.google.com/en//archive/gfs-sosp2003.pdf).

## Problem Scope 
Data collection is increasingly important for companies to stay ahead, however, the large amount prevents them from storing it on a single server. Additionally, accessing large amounts from a distant location can be very slow. A solution to this is creating a distributed system of smaller chunk servers to help store their data. This system will ensure high performance as the client will access data from the chunk server closest to them. A master server ensures that chunk servers are alive through regular heartbeats and always ensures files are duplicated across multiple chunk servers ensuring fault tolerance. In the event a chunk server fails, the replicas will guarantee availability as the client can get the file from another chunk server. The file system would allow reads and atomic append functionality. For any atomic append, consistency is guaranteed by ensuring each replica will have the appended update before confirming with the client.

Features Implemented 
* Replication
* Chunking
* Atomic Append
* Heartbeat
* Read

## Done by 
* Kiran Mohan (1004436)
* David Sioson (1004412)
* Peh Jing Xiang (1004276)
* Ryan Seh (1004669)
* Yi Ern (1004266)
