# load_histogram
Small utility to measure response time of web requests

Configure the settings in the code using the constants at the top
Build the program

The built program takes a command line argument of a end point to test
`./load_histogram http://www.google.com`

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
