package model

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"
)

type Navigation struct{
	Employee []string
}

func TestCache(t *testing.T){

	id := "22"
	data := Navigation{Employee: []string{"Hello", "World"}}
	var output Navigation
	payload, err := json.Marshal(data)
	if err != nil{
		t.Fatalf("error json marshal")
	}
	cache := fmt.Sprintf("%s=%s", id, payload)
	index := strings.Index(cache, "=")
	if err := json.Unmarshal([]byte(cache[index+1:]), &output); err != nil{
		t.Fatalf("error unmashing the json: %s", err)
	}
	log.Println(output)
}

