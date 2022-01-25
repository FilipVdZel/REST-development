package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type date_time struct {
	Time string `json: "time"`
}

func main() {
	h1 := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hallo World\n")
	}

	h2 := func(w http.ResponseWriter, req *http.Request) {
		//fmt.Println("GET parameters were:", req.URL.Query().Get("date_only"))
		// Get and format default time responce
		current_time := time.Now()
		t := current_time.Format("2006-01-02 15:04:05")

		//Check Query paraments for additional formating
		param_time := req.URL.Query().Get("time_only")
		param_data := req.URL.Query().Get("date_only")
		if param_time != "" {
			t = current_time.Format("15:04:05")
		}
		if param_data != "" {
			t = current_time.Format("2006-01-02")
		}

		time := date_time{
			Time: t,
		}

		timeJson, err := json.Marshal(time)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
		}

		w.Header().Set("Content-Type", "application/json")

		w.Write(timeJson)
	}

	http.HandleFunc("/", h1)
	http.HandleFunc("/time", h2)

	http.ListenAndServe(":8080", nil)

}
