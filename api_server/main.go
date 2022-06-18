// Package main implements a server for Metadata API service.
package main

import (
	"encoding/json"
	"flag"
	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"google.golang.org/protobuf/encoding/protojson"
	"io/ioutil"
	"log"
	pb "metadata/protos"
	"net/http"
	"net/mail"
	"net/url"
)

var (
	port = flag.String("port", ":8080", "The server port")
)

type server struct {
	// Each application owns one set of application metadata,
	// and each application is identified by their unique source code path
	// We index by source in memory and allow applications to query by it.
	SourceMap map[string]*pb.MetadataParam
	// We also allow applications under the same company to query by company, and get a list of metadata for
	// all applications. It's stored as a map of sets, which maps company name to a set of source paths.
	CompanyMap map[string]map[string]bool
}

func (s *server) CreateMetadata(w http.ResponseWriter, r *http.Request) {
	// Step 1: read YAML data from payload into Metadata proto
	param := &pb.MetadataParam{}
	yamlBytes, err := ioutil.ReadAll(r.Body)
	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)

	// Step 2: mandate all fields are present
	err = protojson.Unmarshal(jsonBytes, param)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	// Step 3: validate email addresses
	for _, maintainer := range param.Maintainers {
		_, err = mail.ParseAddress(maintainer.GetEmail())
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}
	}

	// Step 4: index and store metadata in memory
	if _, ok := s.CompanyMap[param.GetCompany()]; !ok { // company name got updated
		if s.SourceMap[param.GetSource()] != nil {
			// remove source from the old company key
			oldKey := s.SourceMap[param.GetSource()].GetCompany()
			delete(s.CompanyMap[oldKey], s.SourceMap[param.GetSource()].GetSource())
		}
	}
	// save to company map and source map
	s.SourceMap[param.GetSource()] = param
	if s.CompanyMap[param.GetCompany()] == nil {
		s.CompanyMap[param.GetCompany()] = make(map[string]bool)
	}
	s.CompanyMap[param.GetCompany()][param.GetSource()] = true
	json.NewEncoder(w).Encode(param)
}

func (s *server) GetMetadata(w http.ResponseWriter, r *http.Request) {
	var paramList []*pb.MetadataParam
	q := r.URL.Query()
	source, err := url.QueryUnescape(q.Get("source"))
	company, err := url.QueryUnescape(q.Get("company"))
	title, err := url.QueryUnescape(q.Get("title"))

	if err != nil {
		json.NewEncoder(w).Encode("failed to parse query parameters")
		return
	}
	// retrieve by source
	if source != "" {
		param := s.SourceMap[source]
		if param != nil {
			paramList = append(paramList, param)
		}
		json.NewEncoder(w).Encode(paramList)
		return
	}

	if company == "" {
		json.NewEncoder(w).Encode("please specify source or company")
		return
	}

	// retrieve by company
	for sourcePath := range s.CompanyMap[company] {
		metadata := s.SourceMap[sourcePath]
		if title == "" {
			paramList = append(paramList, metadata)
		} else {
			// narrow down by title if specified
			if metadata.GetTitle() == title {
				paramList = append(paramList, metadata)
			}
		}
	}

	json.NewEncoder(w).Encode(paramList)
}

func main() {
	flag.Parse()
	s := &server{
		SourceMap:  make(map[string]*pb.MetadataParam),
		CompanyMap: make(map[string]map[string]bool),
	}
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/v1", s.GetMetadata).Methods("GET")
	myRouter.HandleFunc("/v1/metadata", s.CreateMetadata).Methods("POST")

	log.Printf("%v", http.ListenAndServe(*port, myRouter))
}
