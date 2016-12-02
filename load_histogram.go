package main

import (
	"fmt"
	"net/http"
	//"bufio"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/robertkrimen/otto"

	"github.com/tejom/load_histogram/collection"
)

var (
	MIN                    float64
	MAX                    float64
	BUCKETS                int
	COUNT                  int
	THREAD                 int
	REQ_ADDRESS            string
	TEST_CLIENT_PERFORMACE bool
	APPEND_RANDOM          string
	DETAILED_LOGGING       bool
	TIME                   string
)

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
	flag.StringVar(&TIME, "time", "", "Set time to run for, default to seconds,(s,m), Will override count setting")

	flag.Parse()

	if REQ_ADDRESS == "quit" {
		os.Exit(1)
	}

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

	if MIN >= MAX {
		fmt.Printf("Invalid values for min and max ( %f >= %f )\n", MIN, MAX)
		os.Exit(1)
	}
	if !(strings.HasPrefix(REQ_ADDRESS, "http://")) {
		fmt.Println("Address requires http://")
		os.Exit(1)
	}

	if TEST_CLIENT_PERFORMACE {
		SetUpClientTesting()
	}

	coll := collection.NewCollection(MIN, MAX, BUCKETS)
	reqChan := make(chan int, THREAD*2+1)
	resultChan := make(chan float64, COUNT)
	done := make(chan bool, 1)

	userCookie := http.Cookie{}

	var wg sync.WaitGroup
	client := &http.Client{
		Transport: &http.Transport{
			/* not sure what is optimal for these
			or just make them cli arguments */
			MaxIdleConns:        5,
			MaxIdleConnsPerHost: 0,
		},
		Timeout: 5 * time.Second, //make this a variable
	}
	programRunTimeStart := time.Now()
	for x := 0; x < THREAD; x++ {
		wg.Add(1)
		fmt.Printf("Adding thread %d\n", x)

		//create one instance of a javascript vm per thread
		jsvm := otto.New()

		go func() {
			fmt.Println("go reqesuest started")
			for r := range reqChan {
				var request_address string
				fmt.Println("start #", r)
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

				if err != nil {
					fmt.Println(err)
					coll.IncrementErr()

					if TEST_CLIENT_PERFORMACE {
						fmt.Println(" aborting client side test...")
					}

				} else {

					totalClientSideTime := 0.0

					//get client side performace
					if TEST_CLIENT_PERFORMACE {
						wg.Add(1)

						totalClientSideTime = RunClientSideTest(res, client, &wg, jsvm)

						fmt.Println("our total client side time was ", totalClientSideTime)
					}

					totalTime := d.Seconds() + totalClientSideTime

					defer res.Body.Close()
					resultChan <- totalTime

					//ensures the tcp connection will be reused

					io.Copy(ioutil.Discard, res.Body)
					ioutil.ReadAll(res.Body)

				}

			}

			defer wg.Done()
		}()
	}

	run := true
	if TIME != "" {
		totalSeconds, err := time.ParseDuration(TIME)
		if err != nil {
			fmt.Println("Invalid time: ", TIME)
			os.Exit(1)
		}
		timer := time.NewTimer(totalSeconds)
		go func() {
			i := 1
			for run {
				reqChan <- i
				//fmt.Println("adding...")
				//time.Sleep(1)
				i += 1
			}
			fmt.Println("stoped??...")
			close(reqChan)
			coll.SetStatTotal(i)
		}()
		go func() {
			//wait to turn timer off without blocking
			<-timer.C
			fmt.Println("Time experired")
			run = false
		}()
	} else {
		//queue up jobs
		for i := 0; i < COUNT; i++ {
			reqChan <- i
		}
		close(reqChan)
	}

	go func() {
		for seconds := range resultChan {
			coll.Add(seconds)
		}
		done <- true //allow all results to be proccessed before continuing
	}()
	wg.Wait()
	totalRunTime := time.Now().Sub(programRunTimeStart)
	coll.SetRunTime(totalRunTime)
	close(resultChan)
	<-done

	coll.PrintGraph()
	coll.CalculateStats()
}
