package main

import (
	"log"
	"net/http"
	wikiapi "widiff/wiki_api"
)

func main() {

	serveMux := http.NewServeMux()
	// serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("Hello WiDiff"))
	// })

	serveMux.HandleFunc("/diff", func(w http.ResponseWriter, r *http.Request) {
		diff, err := wikiapi.GetDiff(wikiapi.TestUrl)
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
