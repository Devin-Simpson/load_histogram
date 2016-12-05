package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

func initFlags() {
	flag.StringVar(&REQ_ADDRESS, "address", "quit", "The web address to load test, if blank, will cancel test")
	flag.Float64Var(&MIN, "min", 0.0, "The minimum response time shown in the histogram")
	flag.Float64Var(&MAX, "max", 2.0, "The maximum response time shown in the histogram")
	flag.IntVar(&BUCKETS, "buckets", 30, "The number of buckets comprising the histogram")
	flag.IntVar(&COUNT, "count", 100, "The number of request jobs")
	flag.IntVar(&THREAD, "thread", 5, "The number of threads to spawn")
	flag.BoolVar(&TEST_CLIENT_PERFORMACE, "testClient", false,
		"Run client side performace test\n\tParse html response and include dependent files in benchmark time")
	flag.StringVar(&APPEND_RANDOM, "paramName", "", "Append given parameter with a unique value")
	flag.BoolVar(&DETAILED_LOGGING, "detailedLogging", false, "Print out detailed logs durring testing")
	flag.StringVar(&TIME, "time", "", "Set time to run for, default to seconds,(s,m), Will override count setting")

	flag.Parse()
}

func printOptions() {
	fmt.Printf("Requests to %s \n", REQ_ADDRESS)
	fmt.Printf("Min bucket time: %f \n", MIN)
	fmt.Printf("Max bucket time: %f \n", MAX)
	fmt.Printf("Buckets: %d \n", BUCKETS)
	if TIME == "" {
		fmt.Printf("Request count: %d \n", COUNT)
	} else {
		fmt.Printf("Run for: %s \n", TIME)
	}
	fmt.Printf("Thread count: %d \n", THREAD)
	fmt.Printf("Testing client: %t \n", TEST_CLIENT_PERFORMACE)
	fmt.Printf("Show detailed logging: %t \n", DETAILED_LOGGING)
}

func checkOptions() error {

	if REQ_ADDRESS == "quit" {
		return errors.New("No address to test given")
	}

	if MIN >= MAX {
		s := fmt.Sprintf("Invalid values for min and max ( %f >= %f )\n", MIN, MAX)
		return errors.New(s)
	}
	if !(strings.HasPrefix(REQ_ADDRESS, "http://") || strings.HasPrefix(REQ_ADDRESS, "https://")) {

		return errors.New("Address requires http:// or https://")
	}

	return nil
}
