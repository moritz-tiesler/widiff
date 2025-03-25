package wikiapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const actionPrefix = "https://en.wikipedia.org/w/api.php?action="

var TestUrl string = "https://en.wikipedia.org/w/api.php?action=compare&format=json&fromtitle=Leipzig_Gewandhaus_Orchestra&fromrev=1279403497&totitle=Leipzig_Gewandhaus_Orchestra&torev=1282274233&difftype=unified&utf8=1&formatversion=2"

var TestRequest CompareRequest = CompareRequest{
	FromTitle: "Leipzig_Gewandhaus_Orchestra",
	FromRevId: "1279403497",
	ToTitle:   "Leipzig_Gewandhaus_Orchestra",
	ToRevId:   "1282274233",
}

type CompareRequest struct {
	FromTitle string
	ToTitle   string
	FromRevId string
	ToRevId   string
}

func (cr *CompareRequest) URL() string {
	title := strings.Replace(cr.FromTitle, " ", "_", -1)

	var b strings.Builder
	b.WriteString(actionPrefix)
	b.WriteString("compare")
	b.WriteString("&format=json")
	b.WriteString(fmt.Sprintf("&fromtitle=%s", title))
	b.WriteString(fmt.Sprintf("&fromrev=%s", cr.FromRevId))
	b.WriteString(fmt.Sprintf("&totitle=%s", title))
	b.WriteString(fmt.Sprintf("&torev=%s", cr.ToRevId))

	b.WriteString(fmt.Sprintf("&difftype=%s", "unified"))
	b.WriteString(fmt.Sprintf("&utf8=%d", 1))
	b.WriteString(fmt.Sprintf("&formatversion=%d", 2))

	return b.String()
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

	var diff CompareResponse
	err = json.Unmarshal(body, &diff)
	if err != nil {
		return nil, err
	}

	return &diff, nil
}

type RecentChangesResponse struct {
	Query Query `json:"query"`
}

type Query struct {
	RecentChanges []RecentChange `json:"recentchanges"`
}

type RecentChange struct {
	Type          string `json:"type"`
	Ns            int    `json:"ns"`
	Title         string `json:"title"`
	PageID        int    `json:"pageid"` // pageid can be used as curid https://en.wikipedia.org/wiki/?curid=31388
	RevID         int    `json:"revid"`
	OldRevID      int    `json:"old_revid"`
	Rcid          int    `json:"rcid"`
	OldLen        int    `json:"oldlen"`
	NewLen        int    `json:"newlen"`
	Timestamp     string `json:"timestamp"`
	ParsedComment string `json:"parsedcomment"`
}

func GetRecentChanges() (*RecentChangesResponse, error) {
	url := "https://en.wikipedia.org/w/api.php?action=query&format=json&list=recentchanges&formatversion=2&rcnamespace=0&rcprop=title%7Ctimestamp%7Cids%7Csizes%7Cparsedcomment&rclimit=500&rctype=edit"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var recent RecentChangesResponse
	err = json.Unmarshal(body, &recent)
	if err != nil {
		return nil, err
	}

	return &recent, nil
}
