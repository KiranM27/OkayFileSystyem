package helper

import "time"

const (
	BASE_URL = "http://localhost"

	// IDs
	MASTER_SERVER_PID = 1

	// Ports
	MASTER_SERVER_PORT = 8080
	CLIENT_START_PORT  = 8060

	// Message Types
	DATA_APPEND        = "DATA_APPEND"      // send chunk data to append
	DATA_COMMIT        = "DATA_COMMIT"      // tell chunk to write
	CREATE_NEW_CHUNK   = "CREATE_NEW_CHUNK" // ask the master to create a new chunk
	DATA_PAD           = "DATA_PAD"         // ask chunk server to pad chunk to 10KB
	HEARTBEAT          = "HEARTBEAT"        // hearbeat message by master.
	KILL_YOURSELF      = "KILL_YOURSELF"    // ask chunk to not respond to any messages.
	REVIVE             = "REVIVE"           // ask chunk to come back up and start responding to messages.
	REPLICATE          = "REPLICATE"        // ask chunk (server) to replicate chunk (the file)
	REP_DATA_REQUEST   = "REP_DATA_REQUEST"
	READ_REQ_TO_MASTER = "READ_REQ_TO_MASTER"
	READ_REQ_TO_CHUNK  = "READ_REQ_TO_CHUNK"

	ACK_APPEND       = "ACK_APPEND"       // chunk server ACK that data to be appened has been received
	ACK_COMMIT       = "ACK_COMMIT"       // chunk server ACK that data has been committed
	ACK_CHUNK_CREATE = "ACK_CHUNK_CREATE" // chunk server ACK that new chunk has been created
	ACK_PAD          = "ACK_PAD"          // master waiting for ACK PAD
	ACK_HEARTBEAT    = "ACK_HEARTBEAT"    // chunk server ack that it is still alive
	ACK_REPLICATION  = "ACK_REPLICATION"  // chunk server ack after it has completed replication from another chunkserver
	ACK_READ_REQ     = "ACK_READ_REQ"     // Response by Master
	REP_DATA_REPLY   = "REP_DATA_REPLY"   // Response to data required for replication
	READ_RES         = "READ_RES"         // Response by Chunk

	// Directories from root i.e.)
	ROOT_DIR      = "OkayFileSystem" // Root dir
	DATA_DIR      = "/data"          // directory where the chunk servers store their data
	TEST_DIR      = "/test"          // directory where files corresponding to testing are stored
	TEST_DATA_DIR = "/test/testData" // directory where data corresponding to testing is stored
	START_TIMES_LOG_FILE = "StartTimes.log"
	END_TIMES_LOG_FILE = "EndTimes.log"

	// Unicode 1 Byte character that is used for padding
	PADDING = "~"
	NULL    = "NULL"

	REP_COUNT = 3 // not tested fully with prev code

	CHUNK_SIZE               = 25
	INITIAL_NUM_CHUNKSERVERS = 5
	DEFAULT_TIMEOUT          = 5 * time.Second
	HEARTBEAT_TIMEOUT        = 10 * time.Second
)

var (
	PORT_PID_MAP = map[int]int{ // Maps ports to pids.
		8080: 1,  // Master Server
		8060: 2,  // client
		8091: 11, // chunkserver 1
		8092: 12, // chunkserver 2
		8093: 13, // chunkserver 3
		8094: 14, //chunkserver 4
		8095: 15, //chunkserver 5
	}
)
