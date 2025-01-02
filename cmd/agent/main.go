package main

import (
	"fmt"
	"runtime/metrics"
)

func main() {
	samples := metrics.All()

	for _, s := range samples {
		fmt.Printf("%v\n", s)
	}
}
