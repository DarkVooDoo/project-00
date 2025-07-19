package model

import (
	"log"
	"regexp"
	"testing"
)

type Navigation struct{
	Employee []string
}

func TestPath(t *testing.T){

	referer := "http://localhost:8000/rendez-vous/45"
	isOk, err := regexp.Match("/rendez-vous/[0-9]+$", []byte(referer))
	if err != nil{
		t.Fatalf("error in the regex: %s", err)
	}
	if !isOk{
		log.Printf("error no match")
	}
}

