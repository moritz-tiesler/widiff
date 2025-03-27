package wikiapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
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
	log.Printf("requesting %s", cReq.URL())
	resp, err := http.Get(cReq.URL())
	if err != nil {
		return nil, fmt.Errorf("error on GET request: %v", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var diff wiki.CompareResponse
	err = json.Unmarshal(body, &diff)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling json: %v", err)
	}

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

func TopDiff(startingFrom time.Time) (string, error) {
	recents, err := GetRecentChanges(wiki.RecentChangeRequest{RcEnd: startingFrom})
	if err != nil {
		log.Printf("could not retrieve recents: %s", err)
		return "", err
	}

	log.Printf("num changes: %d", len(recents.Query.RecentChanges))

	longest := LongestChange(recents.Query.RecentChanges)

	compRequest := wiki.CompareRequest{
		FromTitle: longest.Title,
		ToTitle:   longest.Title,
		FromRevId: strconv.Itoa(longest.OldRevID),
		ToRevId:   strconv.Itoa(longest.RevID),
	}

	diff, err := GetCompare(compRequest)
	if err != nil {
		log.Printf("could not retrieve diff: %s", err)
		return "", err
	}
	parsed, err := ParseDiffText(diff.Compare)
	if err != nil {
		log.Printf("could not parse diff: %s", err)
		return "", err
	}

	return parsed, nil
}
