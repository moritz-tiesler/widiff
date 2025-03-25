package wikiapi

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

var TestUrl string = "https://en.wikipedia.org/w/api.php?action=compare&format=json&fromtitle=Leipzig_Gewandhaus_Orchestra&fromrev=1279403497&totitle=Leipzig_Gewandhaus_Orchestra&torev=1282274233&difftype=unified&utf8=1&formatversion=2"

type CompareRequest struct {
	FromId    int
	ToId      int
	FromTitle int
	FromRevId int
	ToRevId   int
	ToTitle   int
}

type CompareResponse struct {
	Compare Comparison `json:"compare"`
}

type Comparison struct {
	FromID    int    `json:"fromid"`
	FromRevID int    `json:"fromrevid"`
	FromNS    int    `json:"fromns"`
	FromTitle string `json:"fromtitle"`
	ToID      int    `json:"toid"`
	ToRevID   int    `json:"torevid"`
	ToNS      int    `json:"tons"`
	ToTitle   string `json:"totitle"`
	Body      string `json:"body"`
}

func GetDiff(url string) (*CompareResponse, error) {
	log.Printf("requesting %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	log.Println(string(body))

	var diff CompareResponse
	err = json.Unmarshal(body, &diff)
	if err != nil {
		return nil, err
	}
	log.Println(diff)

	return &diff, nil
}
