package clientTest

import (
	"fmt"
	"net/http"
	"time"
	"sync"
	"strings"
	"golang.org/x/net/html"
)

func RunClientSideTest(res *http.Response, client *http.Client, wg *sync.WaitGroup, REQ_ADDRESS string) {

	defer wg.Done()

	htmlParser := html.NewTokenizer(res.Body)

	for {

		nextToken := htmlParser.Next()

		switch {
		case nextToken == html.ErrorToken:
			// End of the document, we're done
			return
		case nextToken == html.StartTagToken:
			t := htmlParser.Token()

			isAnchor := t.Data == "script"
			if isAnchor {

				for _, a := range t.Attr {
					if a.Key == "src" {
						clientSideTime := time.Now()
						var assetUrl string

						//maybe better logic for this is possible?
						if strings.HasPrefix(a.Val, REQ_ADDRESS) {
							assetUrl = a.Val
						} else if strings.HasPrefix(a.Val, "http") {
							assetUrl = a.Val
						} else {
							assetUrl = REQ_ADDRESS + a.Val
						}

						res, err := client.Get(assetUrl)

						if err != nil {
							fmt.Println(err)
						}

						res.Body.Close()

						endClientSideTime := time.Now().Sub(clientSideTime)
						fmt.Println("finished downloading ", a.Val, ", it took ", endClientSideTime)

						break
					}
				}

			}
		}

	}

}