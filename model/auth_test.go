package model

import (
	"os"
	"testing"
)

func TestJson(t *testing.T){
	//cmd := exec.Command("bash", "test.sh", "what")
	//if err := cmd.Run(); err != nil{
	//	log.Fatalf("error command: %s", err)
	//}
	//log.Println("success")
	f, err := os.CreateTemp("Hello.go")
	if err !=  nil{
		t.Fatalf("error creating the file: %s", err)
	}
	defer f.Close()
	
	//myFile, err := os.Open("auth_test.go")
	//if err != nil{
	//	t.Fatalf("error opening the file auth_test: %s", err)
	//}
	//defer myFile.Close()
	//_, err = io.Copy(f, myFile)
	//if err != nil{
	//	t.Fatalf("error copying the file: %s", err)
	//}
	//if err = f.Close(); err != nil{
	//	t.Fatalf("error closing the file: %s", err)
	//}
	//log.Println("Success")
}

