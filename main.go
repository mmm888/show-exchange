package main

import (
	"./mytype"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetData() mytype.Gaitame {
	url := "http://www.gaitameonline.com/rateaj/getrate"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var result mytype.Gaitame
	result.Time = time.Now().Format("2006-01-02 15:04:05")
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		panic(err)
	}
	return result
}

func TopHandler(w http.ResponseWriter, r *http.Request) {
	var data mytype.TestTemplate
	result := GetData()
	data.Table = "<table><tr><th>Name</th><th>Bid</th><th>Ask</th><th>Change(%)</th></tr>"
	funcMap := template.FuncMap{
		"table": func(text string) template.HTML { return template.HTML(text) },
	}
	var open, bid float64
	var change string
	for _, name := range mytype.Exchange_data {
		open, _ = strconv.ParseFloat(result.Quotes[mytype.Exchange_map[name]].Open, 64)
		bid, _ = strconv.ParseFloat(result.Quotes[mytype.Exchange_map[name]].Bid, 64)
		change = fmt.Sprintf("%3.2f", (bid-open)/open*100)
		data.Table = data.Table + "<tr><td>" + result.Quotes[mytype.Exchange_map[name]].Code + "</a></td><td>" + result.Quotes[mytype.Exchange_map[name]].Bid + "</td><td>" + result.Quotes[mytype.Exchange_map[name]].Ask + "</td><td>" + change + "</td></tr>"
	}
	data.Table = data.Table + "</table>"
	data.Main = result.Time
	tmpl, err := template.New("top").Funcs(funcMap).ParseFiles("tmpl/top")
	if err != nil {
		panic(err.Error())
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		panic(err.Error())
	}
}

func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	var data mytype.TestTemplate
	data.Main = "Settings"
	funcMap := template.FuncMap{
		"table": func(text string) template.HTML { return template.HTML(text) },
	}
	tmpl, err := template.New("settings").Funcs(funcMap).ParseFiles("tmpl/settings")
	if err != nil {
		panic(err.Error())
	}
	_ = tmpl.Execute(w, data)
	return
}

func add(id string) []string {
	var judge int
	for _, name := range mytype.Exchange_data {
		if name == id {
			judge = 1
			break
		}
	}
	if judge == 0 {
		mytype.Exchange_data = append(mytype.Exchange_data, id)
	}
	return mytype.Exchange_data
}

func remove(id string) []string {
	var result []string
	for _, name := range mytype.Exchange_data {
		if name != id {
			result = append(result, name)
		}
	}
	return result
}

func CalValue(id, val string) {
	switch val {
	case "Add":
		mytype.Exchange_data = add(id)
	case "Delete":
		mytype.Exchange_data = remove(id)
	}
}

func GetWord(word string) string {
	word = strings.TrimPrefix(word, "[")
	word = strings.TrimSuffix(word, "]")
	return word
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := GetWord(fmt.Sprint(r.Form["id"]))
	val := GetWord(fmt.Sprint(r.Form["Value"]))
	CalValue(id, val)
	http.Redirect(w, r, "/", http.StatusFound)
}

func main() {
	mytype.Exchange_data = append(mytype.Exchange_data, "USDJPY")
	mytype.Exchange_data = append(mytype.Exchange_data, "EURJPY")
	mytype.Exchange_data = append(mytype.Exchange_data, "GBPJPY")
	mytype.Exchange_data = append(mytype.Exchange_data, "CHFJPY")
	mytype.Exchange_data = append(mytype.Exchange_data, "ZARJPY")
	mytype.Exchange_data = append(mytype.Exchange_data, "NZDJPY")
	mytype.Exchange_data = append(mytype.Exchange_data, "AUDJPY")
	mytype.Exchange_data = append(mytype.Exchange_data, "EURUSD")

	http.HandleFunc("/", TopHandler)
	http.HandleFunc("/settings", SettingsHandler)
	http.HandleFunc("/redirect", RedirectHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
