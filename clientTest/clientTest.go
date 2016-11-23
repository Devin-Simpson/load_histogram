package clientTest

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"strings"
	"sync"
	"time"
)

var urlElements map[string]bool
var urlAttributes map[string]bool

// set up the data common to all client side tests before load test is ran
func SetUpClientTesting() {
	//create a map of every element in the DOM that might have a url to follow (maybe better way to do this)
	urlElements = make(map[string]bool)

	urlElements["script"] = true
	urlElements["img"] = true
	urlElements["link"] = true
	urlElements["frame"] = true

	//these are the specific attributes on a DOM element that point to a url which we want to download
	urlAttributes = make(map[string]bool)

	urlAttributes["data-src"] = true
	urlAttributes["src"] = true
	urlAttributes["href"] = true

}

func RunClientSideTest(res *http.Response, client *http.Client, wg *sync.WaitGroup, REQ_ADDRESS string, DETAILED_LOGGING bool) float64 {

	defer wg.Done()

	htmlParser := html.NewTokenizer(res.Body)

	totalClientTime := 0.0

	for {

		nextToken := htmlParser.Next()

		switch {
		case nextToken == html.ErrorToken:
			// End of the document, we're done
			return totalClientTime
		case nextToken == html.StartTagToken:
			token := htmlParser.Token()

			isAnchor := urlElements[token.Data]
			if isAnchor {

				for _, attribute := range token.Attr {
					if urlAttributes[attribute.Key] {
						clientSideTime := time.Now()
						var assetUrl string

						//maybe better logic for this is possible?
						if strings.HasPrefix(attribute.Val, REQ_ADDRESS) {
							assetUrl = attribute.Val
						} else if strings.HasPrefix(attribute.Val, "http") {
							assetUrl = attribute.Val
						} else {
							assetUrl = REQ_ADDRESS + attribute.Val
						}

						res, err := client.Get(assetUrl)

						if err != nil {
							fmt.Println(err)
						}

						res.Body.Close()

						endClientSideTime := time.Now().Sub(clientSideTime)

						totalClientTime = totalClientTime + endClientSideTime.Seconds()

						if DETAILED_LOGGING {
							fmt.Println("finished downloading a ", token.Data, " file: ", attribute.Val, ", it took ", endClientSideTime)
						}

						break
					}
				}

			}
		}

	}

	return totalClientTime

}
