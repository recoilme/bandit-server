package main

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"testing"
)

func TestZero(t *testing.T) {
	log.Println(math.Sqrt((2 * math.Log(float64(1)))))
	fmt.Printf("GOMAXPROCS is %d\n", getGOMAXPROCS())
}

func getGOMAXPROCS() int {
	return runtime.GOMAXPROCS(0)
}
