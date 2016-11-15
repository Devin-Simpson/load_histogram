# load_histogram
Small utility to measure response time of web requests

Configure the settings in the code using the constants at the top
Build the program

The built program takes a command line argument of a end point to test
`./load_histogram -address http://www.google.com`

```
========================================================================================================================================
0.0000 	||[]                                                                                                  	 ||0.0000 ,	 0
0.0667 	||[]                                                                                                  	 ||0.0000 ,	 0
0.1333 	||[xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx]                                                                   	 ||18.1818 ,	 18
```

Example output:
Columns form left to right
Bucket,graph, percentage, total of each request in that bucket.
Each bucket is from the number to the next number. 0.00-0.666 in the above exampe

run ./load_historgram -h for a full list of available options

```
  -address string
    	The web address to load test, if blank, will cancel test (default "quit")
  -buckets int
    	The number of buckets comprising the histogram (default 30)
  -count int
    	The number of request jobs (default 100)
  -max float
    	The maximum response time shown in the histogram (default 2)
  -min float
    	The minimum response time shown in the histogram
  -paramName string
    	Append given parameter with a unique value
  -testClient
    	Run client side performace test
    	Parse html response and include dependent files in benchmark time
  -thread int
    	The number of threads to spawn (default 5)
```
