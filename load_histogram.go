package main

import (
	"fmt"
	"net/http"
	//"bufio"
	"os"
	"sync"
	"time"
	//"io/ioutil"
	"flag"
	"strconv"
	"strings"

	"load_histogram/clientTest"
	"load_histogram/collection"
)

var MIN float64
var MAX float64
var BUCKETS int
var COUNT int
var THREAD int
var REQ_ADDRESS string
var TEST_CLIENT_PERFORMACE bool
var APPEND_RANDOM string
var DETAILED_LOGGING bool

func main() {
	fmt.Printf("%d\n", 0x30)

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

	flag.Parse()

	if REQ_ADDRESS == "quit" {
		os.Exit(1)
	}

	fmt.Printf("Requests to %s \n", REQ_ADDRESS)
	fmt.Printf("Min time: %f \n", MIN)
	fmt.Printf("Max time: %f \n", MAX)
	fmt.Printf("Buckets: %d \n", BUCKETS)
	fmt.Printf("Request count: %d \n", COUNT)
	fmt.Printf("Thread count: %d \n", THREAD)
	fmt.Printf("Testing client: %t \n", TEST_CLIENT_PERFORMACE)
	fmt.Printf("Show detailed logging: %t \n", DETAILED_LOGGING)

	if MIN >= MAX {
		fmt.Printf("Invalid values for min and max ( %f >= %f )\n", MIN, MAX)
		os.Exit(1)
	}
	if !(strings.HasPrefix(REQ_ADDRESS, "http")) {
		fmt.Println("Address requires http://")
		os.Exit(1)
	}

	if TEST_CLIENT_PERFORMACE {
		clientTest.SetUpClientTesting()
	}

	coll := collection.NewCollection(MIN, MAX, BUCKETS)
	reqChan := make(chan int, COUNT)
	resultChan := make(chan float64, COUNT)
	done := make(chan bool, 1)

	userCookie := http.Cookie{}

	var wg sync.WaitGroup

	for x := 0; x < THREAD; x++ {
		wg.Add(1)
		fmt.Printf("Adding thread %d\n", x)
		client := &http.Client{}
		go func() {
			for r := range reqChan {
				var request_address string
				fmt.Println(r)
				if APPEND_RANDOM == "" {
					request_address = REQ_ADDRESS
				} else {
					request_address = REQ_ADDRESS + "?" + APPEND_RANDOM + "=" + strconv.Itoa(r)
				}
				req, err := http.NewRequest("GET", request_address, nil)

				req.AddCookie(&userCookie)
				timeNow := time.Now()
				//res, err := client.Do(req)
				res, err := client.Get(REQ_ADDRESS)

				d := time.Now().Sub(timeNow)
				//htmlData, _ := ioutil.ReadAll(res.Body)
				//fmt.Println(string(htmlData))

				//get client side performace
				if TEST_CLIENT_PERFORMACE {
					wg.Add(1)
					totalClientSideTimeStart := time.Now()

					clientTest.RunClientSideTest(res, client, &wg, REQ_ADDRESS, DETAILED_LOGGING)

					//this is a bad way to sum up client time,
					//we need to add up  individual times to remove the testing overhead
					totalClientSideTimeEnd := time.Now().Sub(totalClientSideTimeStart)
					fmt.Println("our total client side time was ", totalClientSideTimeEnd)
					fmt.Println("backend time", d)
					d = d + totalClientSideTimeEnd
				}

				if err != nil {
					fmt.Println(err)
				} else {
					resultChan <- d.Seconds()
				}
				res.Body.Close()
			}
			defer wg.Done()
		}()
		//		cmd := exec.Command("clear")
		//		cmd.Stdout = os.Stdout
		//		cmd.Run()
		//		coll.printGraph()
	}

	//queue up jobs
	for i := 0; i < COUNT; i++ {
		reqChan <- i
	}
	close(reqChan)
	go func() {
		for seconds := range resultChan {
			coll.Add(seconds)
		}
		done <- true //allow all results to be proccessed before continuing
	}()
	wg.Wait()
	close(resultChan)
	<-done
	coll.PrintGraph()
	coll.CalculateStats()
}
