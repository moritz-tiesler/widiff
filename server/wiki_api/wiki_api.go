package wiki_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"widiff/assert"
	"widiff/wiki"
)

var TestUrl string = "https://en.wikipedia.org/w/api.php?action=compare&format=json&fromtitle=Leipzig_Gewandhaus_Orchestra&fromrev=1279403497&totitle=Leipzig_Gewandhaus_Orchestra&torev=1282274233&difftype=unified&utf8=1&formatversion=2"

var TestRequest wiki.CompareRequest = wiki.CompareRequest{
	FromTitle: "Leipzig_Gewandhaus_Orchestra",
	FromRevId: "1279403497",
	ToTitle:   "Leipzig_Gewandhaus_Orchestra",
	ToRevId:   "1282274233",
}

func GetCompare(cReq wiki.CompareRequest) (*wiki.CompareResponse, error) {
	client := http.DefaultClient
	log.Printf("requesting %s", cReq.URL())
	req, err := http.NewRequest("GET", cReq.URL(), nil)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error on GET request: %v", err)
	}
	if contentType := resp.Header.Get("content-type"); !strings.Contains(contentType, "application/json") {
		log.Println("received html, skipping this page")
		return nil, fmt.Errorf("did not receive json response, received: %s", contentType)
	}

	defer resp.Body.Close()

	var errorBody bytes.Buffer
	var diff wiki.CompareResponse
	r := io.TeeReader(resp.Body, &errorBody)
	err = json.NewDecoder(r).Decode(&diff)

	assert.NoError(
		err,
		"error unmarshaing json",
		map[string]string{
			"header":  fmt.Sprintf("%v", resp.Header),
			"body":    fmt.Sprintf("%s", errorBody.String()),
			"errType": fmt.Sprintf("%T", err),
		},
	)

	return &diff, nil
}

func GetRecentChanges(rcReq wiki.RecentChangeRequest) (*wiki.RecentChangesResponse, error) {
	resp, err := http.Get(rcReq.URL())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var recent wiki.RecentChangesResponse
	err = json.Unmarshal(body, &recent)
	if err != nil {
		return nil, err
	}

	return &recent, nil
}

// TODO: add timestamp for display in frontend
type Diff struct {
	DiffString string
	Comment    string
	User       string
	Size       int
}

// TODO: additionally display change size in bytes
func TopDiff(startingFrom time.Time) (Diff, error) {
	recents, err := GetRecentChanges(wiki.RecentChangeRequest{RcEnd: startingFrom})
	if err != nil {
		log.Printf("could not retrieve recents: %s", err)
		return Diff{}, err
	}

	log.Printf("num changes: %d", len(recents.Query.RecentChanges))

	longest, size := LongestChange(recents.Query.RecentChanges)

	compRequest := wiki.CompareRequest{
		FromTitle: longest.Title,
		ToTitle:   longest.Title,
		FromRevId: strconv.Itoa(longest.OldRevID),
		ToRevId:   strconv.Itoa(longest.RevID),
	}

	diff, err := GetCompare(compRequest)
	if err != nil {
		log.Printf("could not retrieve diff: %s", err)
		return Diff{}, err
	}
	parsed, err := ParseDiffText(diff.Compare)
	if err != nil {
		log.Printf("could not parse diff: %s", err)
		return Diff{}, err
	}

	return Diff{
		DiffString: parsed,
		Comment:    diff.Compare.ToComment,
		Size:       size,
		User:       diff.Compare.FromUser,
	}, nil
}
