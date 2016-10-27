package main

import (
	"fmt"
	"math"
	//"net"
	"net/http"
	//"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
	//"io/ioutil"
	"flag"
)

var MIN float64
var MAX float64
var BUCKETS int
var COUNT int
var THREAD int

type Collection struct {
	min, max, bucketSize, width, height float64
	buckets, count                      int
	coll                                map[float64]float64
	keys                                []float64 //maintain order of coll
}

func NewCollection(min float64, max float64, buckets int) *Collection {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, _ := cmd.Output()
	sizeString := strings.Split(string(out), string(" "))
	width, _ := strconv.ParseFloat(strings.TrimSpace(sizeString[1]), 64)
	fmt.Println(width)

	if buckets == 0 {
		b, _ := strconv.ParseInt(strings.TrimSpace(sizeString[0]), 8, 0)
		buckets = int(b) + 5
	}
	m := make(map[float64]float64)
	keys := make([]float64, buckets+1)
	bucketSize := (max - min) / float64(buckets)
	for i := 0; i <= buckets; i++ {
		// fmt.Println("buckets ", i)
		bucketLimit := bucketSize*float64(i) + min
		m[bucketLimit] = 0.0
		keys[i] = bucketLimit
	}

	c := Collection{
		min:        min,
		max:        max,
		count:      0,
		buckets:    buckets,
		coll:       m,
		bucketSize: bucketSize,
		keys:       keys,
	}

	//get width of terminal for formatting
	c.width = width - 50
	return &c
}

func (c *Collection) add(value float64) {
	b := c.getBucket(value)
	fmt.Println("adding ", value, "to", b)
	c.coll[b]++
	c.count++
}

func (c *Collection) getBucket(value float64) float64 {
	if index := value - math.Mod(value, c.bucketSize); index > c.max {
		return c.max
	} else {
		return index + c.min
	}
	return 0.0
}

func (c *Collection) printGraph() {
	fmt.Println(strings.Repeat("=", int(c.width)+40))
	max := 0.0
	for _, v := range c.coll {
		if v > max {
			max = v
		}
	}
	for _, val := range c.keys {
		v := c.coll[val]
		p := (v / float64(c.count)) * 100.0
		charNum := v / max * 100
		chars := (charNum / 100.0) * c.width
		fmt.Printf("%.4f \t||", val)
		fmt.Printf("[%s]", strings.Repeat("x", int(chars)))
		fmt.Printf("%s", strings.Repeat(" ", int(c.width-chars)))
		fmt.Printf("  \t ||%.4f ,\t %.0f\n", p, v)

	}
	fmt.Printf("%d\n", c.count)
}

func main() {
	fmt.Printf("%d\n", 0x30)

	var REQ_ADDRESS string

	flag.StringVar(&REQ_ADDRESS, "address", "quit", "The web address to load test, if blank, will cancel test")
	flag.Float64Var(&MIN, "min", 0.0, "The minimum response time shown in the histogram")
	flag.Float64Var(&MAX, "max", 2.0, "The maximum response time shown in the histogram")
	flag.IntVar(&BUCKETS, "buckets", 30, "The number of buckets comprising the histogram")
	flag.IntVar(&COUNT, "count", 100, "The number of request jobs")
	flag.IntVar(&THREAD, "thread", 5, "The number of threads to spawn")

	flag.Parse()

	fmt.Printf("Requests to %s \n", REQ_ADDRESS)
	fmt.Printf("Min time: %f \n", MIN)
	fmt.Printf("Max time: %f \n", MAX)
	fmt.Printf("Buckets: %d \n", BUCKETS)
	fmt.Printf("Request count: %d \n", COUNT)
	fmt.Printf("Thread count: %d \n", THREAD)

	if REQ_ADDRESS == "quit" {
		os.Exit(1)
	}

	coll := NewCollection(MIN, MAX, BUCKETS)
	reqChan := make(chan int, COUNT)
	resultChan := make(chan float64, COUNT)

	userCookie := http.Cookie{}

	var wg sync.WaitGroup

	for x := 0; x < THREAD; x++ {
		wg.Add(1)
		fmt.Printf("Adding thread %d\n", x)
		client := &http.Client{}
		go func() {
			for r := range reqChan {
				fmt.Println(r)
				req, err := http.NewRequest("GET", REQ_ADDRESS, nil)
				req.AddCookie(&userCookie)
				timeNow := time.Now()
				res, err := client.Do(req)
				//res , err := http.Get(REQ_ADDRESS)
				//res , err := http.Get(REQ_ADDRESS + "?a" + strconv.Itoa(r))
				d := time.Now().Sub(timeNow)
				//htmlData, _ := ioutil.ReadAll(res.Body)
				//fmt.Println(string(htmlData))
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
			coll.add(seconds)
		}
	}()
	wg.Wait()
	close(resultChan)
	coll.printGraph()
}
