package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type JobPosting struct {
	Title    string
	Company  string
	Location string
	URL      string
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Please visit http://localhost:8080/ to see the results")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// if scraping the listings error, panic as no results will be returned
	jobs, err := scrapeJobListings()
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

func scrapeJobListings() ([]JobPosting, error) {
	var jobs []JobPosting

	listingsURL := getListingsURLFlag()
	urls, err := parseResponseURLs(listingsURL)
	if err != nil {
		return nil, err
	}

	for _, url := range urls {
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		client := &http.Client{}
		resp, err := client.Do(request)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		title := getTitleContents(string(bytes))
		job := getJobFromTitleContents(title, url)
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// set the initial url when running the program
func getListingsURLFlag() string {
	url := flag.String("url", "", "URL for requesting job posting urls")
	flag.Parse()

	if *url == "" {
		log.Fatal("Must supply url of job listings")
	}

	return *url
}

// get the list of job posting urls
func parseResponseURLs(url string) ([]string, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var urls []string
	json.NewDecoder(request.Body).Decode(&urls)

	return urls, nil
}

// extract title tag in the response body with all of the info we need
func getTitleContents(body string) string {
	rgx := regexp.MustCompile(`\<title>(.*?)\</title>`)
	results := rgx.FindStringSubmatch(body)
	return results[1]
}

// split the separate pieces of info returned in the response body title tag
func getJobFromTitleContents(title, url string) JobPosting {
	dashSplit := strings.Split(title, " - ")
	pipeSplit := strings.Split(dashSplit[len(dashSplit)-1], " | ")

	var job JobPosting
	job.Title = strings.Trim(dashSplit[0], " job")
	job.Company = dashSplit[1]
	job.Location = pipeSplit[0]
	job.URL = url

	return job
}
