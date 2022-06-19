// Package main implements a test client that sends requests to Metadata API
package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	writeAddr = flag.String("write_addr", "http://localhost:8080/v1/metadata", "the address to post to")
	readAddr  = flag.String("read_addr", "http://localhost:8080/v1", "the address to get from")
)

func main() {
	flag.Parse()
	client := &http.Client{}

	verifyValidPayloadPersistsSuccessfully(client)
	verifyNotMatchingSourceReturnsNull(client)
	verifyGetTitleReturnsError(client)
	verifyInvalidEmailFailsPersist(client)
	verifyMissingVersionFailsPersist(client)
	verifyQueryByCompanyAndTitle(client)
	verifySameSourceUpdatesLastVersion(client)
	verifySameCompanyReturnsList(client)
	verifyChangeCompanyNameUpdatesDB(client)
}

func verifyValidPayloadPersistsSuccessfully(client *http.Client) {
	b, err := ioutil.ReadFile("valid_example_1.yaml")

	// Step 1: POST the valid payload
	req, err := http.NewRequest("POST", *writeAddr, bytes.NewReader(b))
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)

	log.Printf("test1 create metadata response : %v %v", string(body), err)

	// Step 2: GET the metadata by source
	req, err = http.NewRequest("GET", *readAddr, nil)
	q := req.URL.Query()
	q.Add("source", "https://github.com/random/repo")
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	body, err = ioutil.ReadAll(resp.Body)

	log.Printf("test1 get metadata response: %v %v", string(body), err)
}

func verifyNotMatchingSourceReturnsNull(client *http.Client) {
	// GET no metadata when source doesn't match
	req, err := http.NewRequest("GET", *readAddr, nil)
	q := req.URL.Query()
	q.Add("source", "https://not/stored/repo")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("test2 get metadata response: %v %v", string(body), err)
}

func verifyGetTitleReturnsError(client *http.Client) {
	// GET error message when only title is specified
	req, err := http.NewRequest("GET", *readAddr, nil)
	q := req.URL.Query()
	q.Add("title", "Title only shouldn't work")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("test3 get metadata response: %v %v", string(body), err)
}

func verifyInvalidEmailFailsPersist(client *http.Client) {
	b, err := ioutil.ReadFile("invalid_email.yaml")

	// invalid payload returns error
	req, err := http.NewRequest("POST", *writeAddr, bytes.NewReader(b))
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("test4 create metadata response: %v %v", string(body), err)
}

func verifyMissingVersionFailsPersist(client *http.Client) {
	b, err := ioutil.ReadFile("missing_version.yaml")

	// invalid payload returns error
	req, err := http.NewRequest("POST", *writeAddr, bytes.NewReader(b))
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("test5 create metadata response: %v %v", string(body), err)
}

func verifyQueryByCompanyAndTitle(client *http.Client) {
	b, err := ioutil.ReadFile("valid_example_2.yaml")

	// Step 1: POST the valid payload
	req, err := http.NewRequest("POST", *writeAddr, bytes.NewReader(b))
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("test6 create metadata response: %v %v", string(body), err)

	// Step 2: GET the metadata by company and title
	req, err = http.NewRequest("GET", *readAddr, nil)
	q := req.URL.Query()
	q.Add("company", "Upbound Inc.")
	q.Add("title", "Valid App 2")
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	body, err = ioutil.ReadAll(resp.Body)
	log.Printf("test6 get metadata response: %v %v", string(body), err)
}

func verifySameSourceUpdatesLastVersion(client *http.Client) {
	b, err := ioutil.ReadFile("valid_replace_last.yaml")

	// Step 1: POST the valid payload
	req, err := http.NewRequest("POST", *writeAddr, bytes.NewReader(b))
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("test7 create metadata response: %v %v", string(body), err)

	// Step 2: GET the metadata by source
	req, err = http.NewRequest("GET", *readAddr, nil)
	q := req.URL.Query()
	q.Add("source", "https://github.com/random/repo")
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	body, err = ioutil.ReadAll(resp.Body)
	log.Printf("test7 get metadata response: %v %v", string(body), err)
}

func verifySameCompanyReturnsList(client *http.Client) {
	b, err := ioutil.ReadFile("valid_same_company.yaml")

	// Step 1: POST the valid payload
	req, err := http.NewRequest("POST", *writeAddr, bytes.NewReader(b))
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("test8 create metadata response: %v %v", string(body), err)

	// Step 2: GET the metadata by company
	req, err = http.NewRequest("GET", *readAddr, nil)
	q := req.URL.Query()
	q.Add("company", "Random Inc.")
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	body, err = ioutil.ReadAll(resp.Body)
	log.Printf("test8 get metadata response: %v %v", string(body), err)
}

func verifyChangeCompanyNameUpdatesDB(client *http.Client) {
	b, err := ioutil.ReadFile("valid_different_company.yaml")

	// Step 1: POST the valid payload
	req, err := http.NewRequest("POST", *writeAddr, bytes.NewReader(b))
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("test9 create metadata response: %v %v", string(body), err)

	// Step 2: GET the metadata by old company should not include this one
	req, err = http.NewRequest("GET", *readAddr, nil)
	q := req.URL.Query()
	q.Add("company", "Random Inc.")
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	body, err = ioutil.ReadAll(resp.Body)
	log.Printf("test9 get metadata response: %v %v", string(body), err)

	// Step 3: GET the metadata by new company name should include this one
	req, err = http.NewRequest("GET", *readAddr, nil)
	q = req.URL.Query()
	q.Add("company", "New Random LLC.")
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	body, err = ioutil.ReadAll(resp.Body)
	log.Printf("test9 get metadata response: %v %v", string(body), err)

	// Step 4: Query by source should return the updated metadata
	req, err = http.NewRequest("GET", *readAddr, nil)
	q = req.URL.Query()
	q.Add("source", "https://github.com/random/repo")
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	body, err = ioutil.ReadAll(resp.Body)
	log.Printf("test9 get metadata response: %v %v", string(body), err)
}
