package main

import (
	"log"
	"net/http"
	"strconv"
	wikiapi "widiff/wiki_api"
)

func main() {

	serveMux := http.NewServeMux()
	// serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("Hello WiDiff"))
	// })

	serveMux.HandleFunc("/diff", func(w http.ResponseWriter, r *http.Request) {
		// diff, err := wikiapi.GetDiff(wikiapi.TestUrl)
		diff, err := wikiapi.GetDiff(wikiapi.TestRequest.URL())
		if err != nil {
			log.Printf("could not retrieve diff: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		parsed, err := wikiapi.ParseDiffText(diff.Compare)
		if err != nil {
			log.Printf("could not parse diff: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write([]byte(parsed))
	})

	serveMux.HandleFunc("/diffs", func(w http.ResponseWriter, r *http.Request) {
		recents, err := wikiapi.GetRecentChanges()
		if err != nil {
			log.Printf("could not retrieve recents: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		longest := wikiapi.LongestChange(recents.Query.RecentChanges)

		compRequest := wikiapi.CompareRequest{
			FromTitle: longest.Title,
			ToTitle:   longest.Title,
			FromRevId: strconv.Itoa(longest.OldRevID),
			ToRevId:   strconv.Itoa(longest.RevID),
		}

		diff, err := wikiapi.GetDiff(compRequest.URL())
		if err != nil {
			log.Printf("could not retrieve diff: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		parsed, err := wikiapi.ParseDiffText(diff.Compare)
		if err != nil {
			log.Printf("could not parse diff: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write([]byte(parsed))
	})
	// serveMux.Handle("/view", http.StripPrefix("/view", http.FileServer(http.Dir("./static"))))
	serveMux.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		http.ServeFile(w, r, "./static/index.html")
	})

	http.ListenAndServe(":8080", serveMux)
}
