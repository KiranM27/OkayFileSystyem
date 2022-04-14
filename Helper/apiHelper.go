package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	structs "oks/Structs"
	"strconv"
)

// Destination Port (increased Pointer) --> message.Port[message.Pointer], increase Pointer before sending
func SendMessage(message structs.Message) { // V2 takes in a Message object directly.
	portNo := message.Ports[message.Pointer]
	requestUrl := BASE_URL + ":" + strconv.Itoa(portNo) + "/message"
	messageJSON, _ := json.Marshal(message)
	body := POSTMessage(requestUrl, messageJSON)
	fmt.Println(string(body))
}

func SendRepMsg(repMsg structs.RepMsg, receiver int) {
	portNo := receiver
	requestUrl := BASE_URL + ":" + strconv.Itoa(portNo) + "/replicate"
	messageJSON, _ := json.Marshal(repMsg)
	body := POSTMessage(requestUrl, messageJSON)
	fmt.Println(string(body))
}

func SendReadMsg(readMsg structs.ReadMsg, receiver int) {
	portNo := receiver
	requestUrl := BASE_URL + ":" + strconv.Itoa(portNo) + "/read"
	messageJSON, _ := json.Marshal(readMsg)
	body := POSTMessage(requestUrl, messageJSON)
	fmt.Println(string(body))
}

func POSTMessage(requestURL string, messageJSON []byte) []byte { // a helper function that makes POST requests and returns the body of the reponse
	response, err := http.Post(requestURL, "application/json", bytes.NewBuffer(messageJSON))

	//Handle Error (do not log.Fatal so that it doesnt exit)
	if err != nil {
		fmt.Printf("SendRepMsg: An Error Occured - %v\n", err)
		return []byte(NULL)
	} else {
		defer response.Body.Close()
		//Read the response body (do not log.Fatal so that it doesnt exit)
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("SendRepMsg: An Error Occured at Body - %v\n", err)
		}
		return body
	}
}
