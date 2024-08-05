package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"groupietracker"
)

const portNumber = ":8080"

var tpl *template.Template

func init() {

	tpl = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {

	fmt.Println(fmt.Sprintf("Connecting to port %s", portNumber))
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("templates/assets"))

	mux.Handle("/assets/", http.StripPrefix("/assets", fs))
	mux.HandleFunc("/", generateHandler)
	mux.HandleFunc("/artistDetails.html", artistDetailsHandler)
	log.Fatal(http.ListenAndServe(string(portNumber), mux))
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		groupietracker.HandlePageNotFoundError(w, errors.New("invalid Path"))
		return
	}

	data, err := groupietracker.FetchData("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		http.Error(w, "Error fetching data from API", http.StatusInternalServerError)
		return
	}

	_, err = groupietracker.FetchLocations("https://groupietrackers.herokuapp.com/api/locations")
	if err != nil {
		http.Error(w, "Error fetching data from API", http.StatusInternalServerError)
		return
	}

	_, err = groupietracker.FetchDates("https://groupietrackers.herokuapp.com/api/dates")
	if err != nil {
		http.Error(w, "Error fetching data from API", http.StatusInternalServerError)
		return
	}

	input := r.FormValue("Name")

	// Are we filtering data here by a specific input???
	filtered := make([]groupietracker.Artists, 0)
	for _, obj := range data {

		if strings.Contains(strings.ToLower(obj.Name), strings.ToLower(input)) {
			filtered = append(filtered, obj)
		}

	}

	if len(filtered) == 0 {
		input = ""
	}

	if input != "" {
		// is this being invoked? Are we able to sort the data?
		tpl.ExecuteTemplate(w, "index.html", filtered)
		// groupietracker.RenderTemplate(w, "index.html", filtered)
		return
	}

	// groupietracker.RenderTemplate(w, "index.html", data)
	tpl.ExecuteTemplate(w, "index.html", data)

}

func artistDetailsHandler(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")

	idnum, err := strconv.Atoi(id)
	if err != nil || idnum <= 0 || idnum >= 53 {
		// why idnum >= 53???
		groupietracker.HandlePageNotFoundError(w, err)
		return
	}

	artistData, err := groupietracker.FetchArtistData("https://groupietrackers.herokuapp.com/api/artists/" + id)
	if err != nil {
		http.Error(w, "Error fetching artist data from API", http.StatusInternalServerError)
		return
	}

	relationData, err := groupietracker.FetchRelationData("https://groupietrackers.herokuapp.com/api/relation/" + id)
	if err != nil {
		http.Error(w, "Error fetching relation data from API", http.StatusInternalServerError)
		return
	}

	tpl.ExecuteTemplate(w, "artistDetails.html", struct {
		Artists  groupietracker.Artists
		Relation groupietracker.Relations
	}{
		Artists:  artistData,
		Relation: relationData,
	})

}
