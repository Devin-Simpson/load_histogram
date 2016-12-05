package main

import (
	"fmt"
	"net/http"
	//"bufio"
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"

	"sync"
	"syscall"
	"time"

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

	initFlags()

	startErr := checkOptions()
	if startErr != nil {
		fmt.Println(startErr.Error())
		os.Exit(1)
	}

	if TEST_CLIENT_PERFORMACE {
		SetUpClientTesting()
	}
	printOptions()

	coll := collection.NewCollection(MIN, MAX, BUCKETS)
	reqChan := make(chan int, THREAD*2+1)
	resultChan := make(chan float64, COUNT)
	done := make(chan bool, 1)
	breakLoop := false
	run := true //set incase time is set

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

	//set up handling signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan

		if TIME != "" {
			run = false
		}
		breakLoop = true
		drainedReq := 0
		for x := range reqChan {
			drainedReq = drainedReq + 1
			fmt.Println(drainedReq, "#", x)
		}
		fmt.Println("drained requests", drainedReq)

	}()

	for x := 0; x < THREAD; x++ {
		wg.Add(1)
		fmt.Printf("Adding thread %d\n", x)

		//create one instance of a javascript vm per thread
		jsvm := otto.New()

		go func() {
			fmt.Println("go reqesuest started")
			for r := range reqChan {
				fmt.Println("start #", r)

				var request_address string

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

				if breakLoop {
					break
				}

			}

			defer wg.Done()
		}()
	}

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
				if breakLoop {
					break
				}
				reqChan <- i
				i += 1
				fmt.Println("adding to que #", i)
			}
			fmt.Println("stoped??...")
			close(reqChan)

		}()
		go func() {
			<-timer.C //wait for timer t expire without blocking
			fmt.Println("Time experired")
			run = false
		}()
	} else {
		//queue up jobs
		for i := 0; i < COUNT; i++ {
			if breakLoop {
				break
			}
			reqChan <- i

		}
		close(reqChan)
	}

	go func() {
		count := 0
		for seconds := range resultChan {
			coll.Add(seconds)
			count += 1
		}
		coll.SetStatTotal(count)
		done <- true //allow all results to be proccessed before continuing
	}()
	fmt.Println("wg wait?")
	wg.Wait()
	totalRunTime := time.Now().Sub(programRunTimeStart)
	coll.SetRunTime(totalRunTime)
	close(resultChan)
	fmt.Println("get to done?")
	<-done

	coll.PrintGraph()
	coll.CalculateStats()
}
