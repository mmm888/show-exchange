package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var exchange_data []string
var exchange_map = map[string]int{"GBPNZD": 0, "CADJPY": 1, "GBPAUD": 2, "AUDJPY": 3, "AUDNZD": 4, "EURCAD": 5, "EURUSD": 6, "NZDJPY": 7, "USDCAD": 8, "EURGBP": 9, "GBPUSD": 10, "ZARJPY": 11, "EURCHF": 12, "CHFJPY": 13, "AUDUSD": 14, "USDCHF": 15, "EURJPY": 16, "GBPCHF": 17, "EURNZD": 18, "NZDUSD": 19, "USDJPY": 20, "EURAUD": 21, "AUDCHF": 22, "GBPJPY": 23}

type Exchange_db struct {
	time string
	open string
	bid  string
	ask  string
	high string
	low  string
}

type Gaitame struct {
	Quotes []struct {
		Code string `json:"currencyPairCode"`
		Open string `json:"open"`
		Bid  string `json:"bid"`
		Ask  string `json:"ask"`
		High string `json:"high"`
		Low  string `json:"low"`
	} `json:"quotes"`
	time string
}

type testTemplate struct {
	Main  string
	Sub   string
	Graph string
	Date  string
	Table string
}

type User struct {
	user string
	pass string
}

func GetData() Gaitame {
	url := "http://www.gaitameonline.com/rateaj/getrate"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var result Gaitame
	result.time = time.Now().Format("2006-01-02 15:04:05")
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		panic(err)
	}
	return result
}

func TopHandler(w http.ResponseWriter, r *http.Request) {
	var data testTemplate
	result := GetData()
	data.Table = "<table><tr><th>Name</th><th>Bid</th><th>Ask</th><th>Change(%)</th></tr>"
	funcMap := template.FuncMap{
		"table": func(text string) template.HTML { return template.HTML(text) },
	}
	var open, bid float64
	var change string
	for _, name := range exchange_data {
		open, _ = strconv.ParseFloat(result.Quotes[exchange_map[name]].Open, 64)
		bid, _ = strconv.ParseFloat(result.Quotes[exchange_map[name]].Bid, 64)
		change = fmt.Sprintf("%3.2f", (bid-open)/open*100)
		data.Table = data.Table + "<tr><td>" + result.Quotes[exchange_map[name]].Code + "</a></td><td>" + result.Quotes[exchange_map[name]].Bid + "</td><td>" + result.Quotes[exchange_map[name]].Ask + "</td><td>" + change + "</td></tr>"
	}
	data.Table = data.Table + "</table>"
	data.Main = result.time
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
	var data testTemplate
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

func add(exchange_data []string, id string) []string {
	var judge int
	for _, name := range exchange_data {
		if name == id {
			judge = 1
			break
		}
	}
	if judge == 0 {
		exchange_data = append(exchange_data, id)
	}
	return exchange_data
}

func remove(exchange_data []string, id string) []string {
	var result []string
	for _, name := range exchange_data {
		if name != id {
			result = append(result, name)
		}
	}
	return result
}

func CalValue(id, val string) {
	switch val {
	case "Add":
		exchange_data = add(exchange_data, id)
	case "Delete":
		exchange_data = remove(exchange_data, id)
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
	exchange_data = append(exchange_data, "USDJPY")
	exchange_data = append(exchange_data, "EURJPY")
	exchange_data = append(exchange_data, "GBPJPY")
	exchange_data = append(exchange_data, "CHFJPY")
	exchange_data = append(exchange_data, "ZARJPY")
	exchange_data = append(exchange_data, "NZDJPY")
	exchange_data = append(exchange_data, "AUDJPY")
	exchange_data = append(exchange_data, "EURUSD")

	http.HandleFunc("/", TopHandler)
	http.HandleFunc("/settings", SettingsHandler)
	http.HandleFunc("/redirect", RedirectHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
