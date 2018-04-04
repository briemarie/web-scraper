package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type JobDetails struct {
	Title    string
	Company  string
	Location string
	URL      string
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// if scraping the listings error, panic as no results will be returned
	jobs, err := scrapeJobListings(r.URL.String())
	if err != nil {
		panic(err)
	}

	for _, job := range jobs {
		jsonJob, err := json.Marshal(job)
		// if one listing can not be marshalled, log the error, but try the next
		if err != nil {
			log.Printf("Error marshalling json: %v", err)
			continue
		}

		w.Write(jsonJob)
	}
}

// return response from a get request
func getResponse(url string) (*http.Response, error){
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// get all of the jobs' details concurrently
func scrapeJobListings(url string) ([]JobDetails, error) {
	var jobs []JobDetails

	urls, err := parseResponseURLs(url)
	if err != nil {
		return nil, err
	}

	chJobs := make(chan JobDetails)
	chFinished := make(chan bool)

	for _, url := range urls {
		go scrapeJobDetails(url, chJobs, chFinished)
	}

	for c := 0; c < len(urls); {
		select {
		case job := <-chJobs:
			jobs = append(jobs, job)
		case <-chFinished:
			c++
		}
	}

	close(chJobs)

	return jobs, nil
}

// get the list of job posting urls from the request url
func parseResponseURLs(url string) ([]string, error) {
	resp, err := getResponse(url)
	if err != nil {
		return nil, err
	}

	var urls []string
	json.NewDecoder(resp.Body).Decode(&urls)

	return urls, nil
}

// scrape the details of the job
// (using channels so that other jobs are scraped concurrently)
func scrapeJobDetails(url string, ch chan JobDetails, chFinished chan bool) error {
	resp, err := getResponse(url)
	if err != nil {
		return err
	}

	defer func() {
		chFinished <- true
	}()

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	title := scrapeTitleTag(string(bytes))
	ch <- getJobDetails(title, url)

	return nil
}

// scrape the title tag from response body (has the job details we need)
func scrapeTitleTag(body string) string {
	rgx := regexp.MustCompile(`\<title>(.*?)\</title>`)
	results := rgx.FindStringSubmatch(body)
	return results[1]
}

// separate the job details returned in the response body title tag
func getJobDetails(title, url string) JobDetails {
	dashSplit := strings.Split(title, " - ")
	pipeSplit := strings.Split(dashSplit[len(dashSplit)-1], " | ")

	var job JobDetails
	job.Title = strings.Trim(dashSplit[0], " job")
	job.Company = dashSplit[1]
	job.Location = pipeSplit[0]
	job.URL = url

	return job
}
