# Welcome to my Indeed Job Posting web scraper

### Requirements
1. Go (I have v1.10 installed)

### To use:
By downloading the web_scraper.go file, then:
1. From the command line, navigate into the directory where you downloaded the file
2. Run `go build web-scraper.go`

Or, by downloading the web-scraper binary file, then:


To execute the program, run `./web-scraper -url={url of job posting urls}`

When prompted, view the results at http://localhost:8080/


### Alternative methods
Should the web scraper fail to work, you can at least see it in action by replacing lines 50-54 of web_scraper.go with the following slice of urls:

urls := []string{
	"http://www.indeed.com/viewjob?jk=8cfd54301d909668",
	"http://www.indeed.com/viewjob?jk=b17c354e3cabe4f1",
	"http://www.indeed.com/viewjob?jk=38123d02e67210d9",
}

You can then run the file either by repeating step 3 above and then running `./web-scraper` or simply run the file with `go run web-scraper.go`
