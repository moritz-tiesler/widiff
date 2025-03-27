package main

import (
	"log"
	"net/http"
	"time"
	"widiff/db"
	wikiapi "widiff/wiki_api"
)

func main() {

	db.TestDb()

	topDiff := ""

	go func() {
		startingFrom := time.Now().Add(-1 * time.Minute)
		newTopDiff, err := wikiapi.TopDiff(startingFrom)
		if err != nil {
			log.Printf("failed to initialize diff: %s", err)
		}
		topDiff = newTopDiff

		ticker := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-ticker.C:
				startingFrom = time.Now().Add(-1 * time.Minute)
				newTopDiff, err = wikiapi.TopDiff(startingFrom)
				if err != nil {
					break
				}
				log.Println("top diff updated")
				topDiff = newTopDiff
			}
		}
	}()

	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/diff", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(topDiff))
	})
	serveMux.Handle("/view/", http.StripPrefix("/view/", http.FileServer(http.Dir("./static"))))

	http.ListenAndServe(":8080", serveMux)
}
