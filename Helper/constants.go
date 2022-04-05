package helper

const (
	BASE_URL = "http://localhost"

	// IDs
	MASTER_SERVER_PID = 1

	// Ports
	MASTER_SERVER_PORT = 8080

	// Message Types
	DATA_APPEND   = "DATA_APPEND"   // send chunk data to append
	DATA_COMMIT   = "DATA_COMMIT"   // tell chunk to write
	DATA_PAD      = "DATA_PAD"      // ask chunk server to pad chunk to 10KB
	HEARTBEAT     = "HEARTBEAT"     // hearbeat message by master.
	KILL_YOURSELF = "KILL_YOURSELF" // ask chunk to not respond to any messages.
	REVIVE        = "REVIVE"        // ask chunk to come back up and start responding to messages.
	REPLICATE     = "REPLICATE"     // ask chunk (server) to replicate chunk (the file)
	REP_DATA_REQUEST= "REP_DATA_REQUEST"
	REP_DATA_REPLY= "REP_DATA_REPLY"
	
	ACK_APPEND       = "ACK_APPEND"       // chunk server ACK that data to be appened has been received
	ACK_COMMIT       = "ACK_COMMIT"       // chunk server ACK that data has been committed
	ACK_CHUNK_CREATE = "ACK_CHUNK_CREATE" // chunk server ACK that new chunk has been created
	ACK_PAD          = "ACK_PAD"          // master waiting for ACK PAD
	ACK_HEARTBEAT    = "ACK_HEARTBEAT"    // chunk server ack that it is still alive
	ACK_REPLICATION  = "ACK_REPLICATION"  // chunk server ack after it has completed replication from another chunkserver

	CREATE_NEW_CHUNK = "CREATE_NEW_CHUNK" // ask the master to create a new chunk

	// Directories from root i.e.)
	ROOT_DIR      = "OkayFileSystem" // Root dir
	DATA_DIR      = "/data"          // directory where the chunk servers store their data
	TEST_DIR      = "/test"          // directory where files corresponding to testing are stored
	TEST_DATA_DIR = "/test/testData" // directory where data corresponding to testing is stored

	// Unicode 1 Byte character that is used for padding
	PADDING = "~"

	REP_COUNT = 3 // not tested fully with prev code 
)

var (
	PORT_PID_MAP = map[int]int{ // Maps ports to pids.
		8080: 1, // Master Server
		8081: 2, // Chunk Server 1
		8082: 3, // Chunk Server 2
		8083: 4, // Chunk Server 3
		8084: 5, // Chunk Server 4
		8085: 6, // Chunk Server 5
		8086: 7, // Client 1
	}
)
