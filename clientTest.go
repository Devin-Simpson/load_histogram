package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/robertkrimen/otto"
)

var urlElements map[string]bool
var urlAttributes map[string]bool

//create a seperate variable for javascript, as we want to execute the javascript (the inline block and downloaded JS)
var jsTag string = "script"

// set up the data common to all client side tests before load test is ran
func SetUpClientTesting() {
	//create a map of every element in the DOM that might have a url to follow (maybe better way to do this)
	urlElements = make(map[string]bool)

	urlElements["img"] = true
	urlElements["link"] = true
	urlElements["frame"] = true

	//these are the specific attributes on a DOM element that point to a url which we want to download
	urlAttributes = make(map[string]bool)

	urlAttributes["data-src"] = true
	urlAttributes["src"] = true
	urlAttributes["href"] = true

}

func RunClientSideTest(res *http.Response, client *http.Client, wg *sync.WaitGroup, jsvm *otto.Otto) float64 {

	defer wg.Done()

	htmlParser := html.NewTokenizer(res.Body)

	totalClientTime := 0.0

	for {

		nextToken := htmlParser.Next()
		token := htmlParser.Token()
		//fmt.Println(token , nextToken)
		switch {
		case nextToken == html.ErrorToken:
			// End of the document, we're done
			return totalClientTime
		case nextToken == html.StartTagToken || nextToken == html.SelfClosingTagToken:
			// case where we are downloading a non javascript file
			isAnchor := urlElements[token.Data]
			if isAnchor {

				for _, attribute := range token.Attr {
					if urlAttributes[attribute.Key] {

						clientSideTime := time.Now()
						var assetUrl = AssetURL(attribute.Val, REQ_ADDRESS)
						downloadedElement, err := client.Get(assetUrl)

						if err != nil {
							fmt.Println(err)
						}

						downloadedElement.Body.Close()
						var logMessage = "finished downloading a " + token.Data + " file: " + attribute.Val + ", it took "
						totalClientTime = LogElapstedTime(clientSideTime, totalClientTime, logMessage, DETAILED_LOGGING)

						break
					}
				}

			} else if token.Data == jsTag {

				for _, attribute := range token.Attr {

					//case where javascript needs to be downloaded first
					if urlAttributes[attribute.Key] {

						clientSideTimeDownLoad := time.Now()
						var assetUrl = AssetURL(attribute.Val, REQ_ADDRESS)
						script, err := client.Get(assetUrl)

						if err != nil {
							fmt.Println(err)
						}

						var logMessageDownload = "downloaded some JS! from " + attribute.Val + " it took "

						totalClientTime = LogElapstedTime(clientSideTimeDownLoad, totalClientTime, logMessageDownload, DETAILED_LOGGING)

						//now execute the javascript and record time to execute + time to download
						if script != nil {

							totalClientTimeJS := time.Now()
							jsvm.Run(script)
							var logMessageJS = "ran some JS! it took "

							totalClientTime = LogElapstedTime(totalClientTimeJS, totalClientTime, logMessageJS, DETAILED_LOGGING)

						} else {
							fmt.Println("ERROR, No js code to execute")
						}

						script.Body.Close()
					}

				}

			}

		}

	}

}

func AssetURL(url string, REQ_ADDRESS string) string {

	var assetUrl string
	//maybe better logic for this is possible?
	if strings.HasPrefix(url, REQ_ADDRESS) {
		assetUrl = url
	} else if strings.HasPrefix(url, "http") {
		assetUrl = url
	} else {
		assetUrl = REQ_ADDRESS + url
	}

	return assetUrl
}

func LogElapstedTime(clientSideTime time.Time, totalClientTime float64, logMessage string, DETAILED_LOGGING bool) float64 {

	endClientSideTime := time.Now().Sub(clientSideTime)
	totalClientTime = totalClientTime + endClientSideTime.Seconds()

	if DETAILED_LOGGING {
		fmt.Println(logMessage, endClientSideTime)
	}

	return totalClientTime
}
