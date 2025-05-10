package model

import (
	"log"
	"testing"
	"time"
)

func TestTime(t *testing.T){
    forwardDate := time.Now().Add(time.Hour * 2)
    result := time.Until(forwardDate)

    log.Printf("Difference: %f", result.Minutes())

}

