# Metadata API Server

The [Go][] implementation of a RESTful API server for application metadata.

POST to endpoint "/v1/metadata" with a YAML payload will persist the application metadata in memory.

GET to endpoint "/v1" with query parameter "source" retrieves the metadata for specified source path.

GET to endpoint "/v1" with query parameter "company" retrieves a list of metadata for all applications belonging to the 
company. Further specifying "title" will narrow down the search.

## Prerequisites

- Install **[Go][]**: https://go.dev/doc/install

  Verify with
    ```
    go version
    ```
- Install proto compiler
    ```
    brew reinstall protobuf
    ```
    For non-mac systems, use other installer such as "apt get install"

  Verify with 
    ```
    protoc --version
    ```
- Configure proto for go
    ```
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
    export PATH="$PATH:$(go env GOPATH)/bin"
    ```

## Design metadata API
Without use cases showing otherwise, each application should own one set of application metadata, and each application 
is identified by their unique source code path.
One company could have multiple applications, and we allow applications to query by company, and get a list of metadata
for all applications belonging to this company.

As a result, we design 2 in-memory maps:
- The first map maps source path to metadata and allows applications to query by it.

    ```
    SourceMap map[string]*pb.MetadataParam
    ```

- The second map is a map of sets, which indexes by company name. To avoid storing duplicate metadata twice, the set 
only contains source paths, which can be used to look up metadata from the first map.
    ```
    CompanyMap map[string]map[string]bool
    ```
With this design, HTTP GET from applications is fast, regardless of query parameters: Querying by source path is a 
constant operation, from the SourceMap lookup.
Querying by company doesn’t require iterating through the SourceMap, because the CompanyMap is kept updated on every 
write (HTTP POST). Detailed sequence is described below:

### CreateMetadata

- Read YAML data from payload into Metadata proto
- Mandate all fields are present
- Validate email addresses
- Index and store metadata in memory
  - If company name changed, remove source path under the old company key
  - Save metadata to company map and source map

### GetMetadata
- Parse query parameters
- If source path is specified, retrieve metadata by source path
- If neither source path or company is specified, return error to application explaining the required query parameters
- Retrieve by company, and if application title is specified, narrow down the search by title


### Open Source Libraries

```console
google/protobuf
```
- Schema for in memory storage (easily extensible to database)

```
github.com/gorilla/mux
```
- Handles http connection and routes endpoints 
- Parses query parameters and writes response

## Test metadata API
#### Generate proto

```
protoc --go_out=. --go_opt=paths=source_relative protos/metadata.proto
```
Verify that metadata.pb.go is generated successfully

#### Run server
```
go run api_server/main.go
```

#### Send test request to server
```
go run test_client/main.go
```

### The test client includes 9 test cases

 1. Test 1: a POST with valid payload followed by a GET with valid query parameter, to verify a valid payload 
should persist successfully 
 2. Test 2: a GET with non matching source path passed in the query parameter should return Null
 3. Test 3: a GET with just the title in the query parameter should return error
 4. Test 4: a POST with invalid email should return error\
 5. Test 5: a POST with missing version should return error
 6. Test 6: a GET with both company and title in the query params should return matching metadata
 7. Test 7: a POST with payload having source path already stored in memory, followed by a GET, to verify metadata is 
updated
 8. Test 8: a GET with company as the query parameter should return a list of metadata belonging to the company
 9. Test 9: a POST with a different company name but the same source path followed by a GET with the original company 
name, a GET with the new company name, and a GET with the source path, to verify that the old company no longer 
contains this metadata, the metadata belongs to the new company, and the source path maps to the updated metadata

[Go]: https://golang.org
