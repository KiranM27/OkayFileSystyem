package helper

import (
	"fmt"
	"os/exec"
)

var ps, err = exec.LookPath("powershell.exe")
func RunClient(id int, portNumber int, sourceFilename string, OFSFilename string ){
	cmd := exec.Command(ps,"Start-Process" ,"PowerShell.exe","-ArgumentList", 
	`"-noexit",`,
	`"-command 
	[console]::WindowWidth=50; [console]::WindowHeight=30; [console]::BufferWidth=[console]::WindowWidth; 
	go run .\Client`+ fmt.Sprintf(` -i %v -p %v -s ..\test\testData\%v -d %v"`,id, portNumber, sourceFilename, OFSFilename))
	cmd.Start()
}

func RunChunkServer( nodePid int, portNo int){
	cmd := exec.Command(ps,"Start-Process" ,"PowerShell.exe","-ArgumentList",`"-noexit",` ,
	`"-noexit",`, // add this so that when code terminates the window stays open. must be exactly the same as what i put here, else the command fails.
	`"-command 
	[console]::WindowWidth=50; [console]::WindowHeight=30; [console]::BufferWidth=[console]::WindowWidth; 
	go run .\ChunkServer`+ fmt.Sprintf(` -i %v -p %v"`,nodePid,portNo))
	cmd.Start()
}
func RunMaster(){
	cmd := exec.Command(ps,"Start-Process" ,"PowerShell.exe","-ArgumentList",`"-noexit",` , `"-command go run .\Master"`)
	cmd.Start()
}