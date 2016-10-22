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
)

const MIN = 0.0
const MAX = 2
const BUCKETS = 30
const COUNT = 100
const thread = 5

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
		bucketLimit := bucketSize * float64(i)
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
	fmt.Println("adding ", value)
	b := c.getBucket(value)
	c.coll[b]++
	c.count++
}

func (c *Collection) getBucket(value float64) float64 {
	if index := value - math.Mod(value, c.bucketSize); index > c.max {
		return c.max
	} else {
		return index
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
	REQ_ADDRESS := os.Args[1]
	fmt.Printf("Requests to %s \n", REQ_ADDRESS)
	coll := NewCollection(MIN, MAX, BUCKETS)
	req_chan := make(chan int, COUNT)

	userCookie := http.Cookie{}

	var wg sync.WaitGroup

	for x := 0; x < thread; x++ {
		wg.Add(1)
		fmt.Printf("Adding thread %d\n", x)
		client := &http.Client{}
		go func() {
			for r := range req_chan {
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
					coll.add(d.Seconds())
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
		req_chan <- i
	}
	close(req_chan)
	wg.Wait()

	coll.printGraph()
}
