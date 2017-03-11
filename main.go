package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type testTemplate struct {
	Main  string
	Sub   string
	Graph string
	Date  string
	Table string
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Go Webapp\n"))
}

type Exchange_db struct {
	time string
	open string
	bid  string
	ask  string
	high string
	low  string
}

func DataHandler(w http.ResponseWriter, r *http.Request) {
	var judge int
	vars := mux.Vars(r)
	for _, n := range exchange_data {
		if n == vars["name"] {
			judge = 1
		}
	}
	if judge == 0 {
		fmt.Fprintf(w, "404 not found.")
	} else {
		var data testTemplate
		data.Main = vars["name"]
		data.Table = "<table><tr><th>Date</th><th>Open</th><th>Bid</th><th>Ask</th><th>High</th><th>Low</th></tr>"
		funcMap := template.FuncMap{
			"graph": func(text string) template.HTML { return template.HTML(text) },
			"table": func(text string) template.HTML { return template.HTML(text) },
		}

		db, err := sql.Open("mysql", "project:YhVd72nv@/project")
		if err != nil {
			panic(err.Error())
		}
		defer db.Close()

		dbformat := "\"%Y-%m-%d\""
		dbname := vars["name"]
		//today := "\"" + time.Now().Format("2006-01-02") + "\""
		yesterday := "\"" + time.Now().Add(-24*time.Hour).Format("2006-01-02") + "\""
		rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s where date_format(time, %s) IN( %s) ORDER BY time ASC", dbname, dbformat, yesterday))
		if err != nil {
			panic(err.Error())
		}
		defer rows.Close()

		var ex_db [500]Exchange_db
		var ex_png string
		var count int
		for rows.Next() {
			err = rows.Scan(&ex_db[count].time, &ex_db[count].open, &ex_db[count].bid, &ex_db[count].ask, &ex_db[count].high, &ex_db[count].low)
			count++
			if err != nil {
				panic(err.Error())
			}
		}

		num := count
		pts := make(plotter.XYs, count)
		var X, Y float64
		var tmp string
		for {
			tmp = strings.Replace(ex_db[count-num].time[11:16], " ", "", -1)
			tmp = strings.Replace(tmp, "-", "", -1)
			tmp = strings.Replace(tmp, ":", "", -1)
			X, _ = strconv.ParseFloat(tmp, 64)
			Y, _ = strconv.ParseFloat(ex_db[count-num].bid, 64)
			pts[count-num].X = X
			pts[count-num].Y = Y
			num--
			data.Table = data.Table + "<tr><td>" + ex_db[num].time + "</td><td>" + ex_db[num].open + "</td><td>" + ex_db[num].bid + "</td><td>" + ex_db[num].ask + "</td><td>" + ex_db[num].high + "</td><td>" + ex_db[num].low + "</td></tr>"
			if num == 0 {
				break
			}
		}
		data.Table = data.Table + "</table>"

		now := time.Now().Format("01-02-15:04")
		p, _ := plot.New()
		p.Title.Text = vars["name"] + " graph"
		p.X.Label.Text = "TIME"
		p.Y.Label.Text = "BID"
		if err := plotutil.AddLinePoints(p, "RATE", pts); err != nil {
			panic(err)
		}
		ex_png = "images/" + "graph" + now + ".png"
		if err := p.Save(5*vg.Inch, 5*vg.Inch, ex_png); err != nil {
			panic(err)
		}
		data.Date = now

		tmpl, err := template.New("detail").Funcs(funcMap).ParseFiles("tmpl/detail")
		if err != nil {
			panic(err.Error())
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			panic(err.Error())
		}

	}
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
		data.Table = data.Table + "<tr><td><a href=\"/detail/" + result.Quotes[exchange_map[name]].Code + "\">" + result.Quotes[exchange_map[name]].Code + "</a></td><td>" + result.Quotes[exchange_map[name]].Bid + "</td><td>" + result.Quotes[exchange_map[name]].Ask + "</td><td>" + change + "</td></tr>"
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

func checkAuth(r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	if ok == false {
		return false
	}
	account := UserPass()
	for _, user := range account {
		if user.user == username && user.pass == password {
			return true
		}
	}
	return false
}

func SecretHandler(w http.ResponseWriter, r *http.Request) {
	var data testTemplate
	data.Main = "Authentication"
	funcMap := template.FuncMap{
		"table": func(text string) template.HTML { return template.HTML(text) },
	}
	tmpl, err := template.New("auth").Funcs(funcMap).ParseFiles("tmpl/auth")
	if err != nil {
		panic(err.Error())
	}
	_ = tmpl.Execute(w, data)
	return
}

func GetWord(word string) string {
	word = strings.TrimPrefix(word, "[")
	word = strings.TrimSuffix(word, "]")
	return word
}

func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := GetWord(fmt.Sprint(r.Form["id"]))
	password := GetWord(fmt.Sprint(r.Form["pass"]))
	r.SetBasicAuth(username, password)
	if checkAuth(r) == false {
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
		return
	}
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

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := GetWord(fmt.Sprint(r.Form["id"]))
	val := GetWord(fmt.Sprint(r.Form["Value"]))
	CalValue(id, val)
	http.Redirect(w, r, "/top", http.StatusFound)
}

var exchange_data []string
var exchange_map = map[string]int{"GBPNZD": 0, "CADJPY": 1, "GBPAUD": 2, "AUDJPY": 3, "AUDNZD": 4, "EURCAD": 5, "EURUSD": 6, "NZDJPY": 7, "USDCAD": 8, "EURGBP": 9, "GBPUSD": 10, "ZARJPY": 11, "EURCHF": 12, "CHFJPY": 13, "AUDUSD": 14, "USDCHF": 15, "EURJPY": 16, "GBPCHF": 17, "EURNZD": 18, "NZDUSD": 19, "USDJPY": 20, "EURAUD": 21, "AUDCHF": 22, "GBPJPY": 23}

type User struct {
	user string
	pass string
}

func UserPass() [5]User {
	db, err := sql.Open("mysql", "project:YhVd72nv@/project")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	rows, err := db.Query("SELECT user,password FROM root")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	var account [5]User
	var count int
	for rows.Next() {
		err = rows.Scan(&account[count].user, &account[count].pass)
		count++
		if err != nil {
			panic(err.Error())
		}
	}
	return account
}

func InDataBase() {
	var result Gaitame
	result = GetData()
	db, err := sql.Open("mysql", "project:YhVd72nv@/project")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	stmtIns, _ := db.Prepare("SELECT Time FROM TEST")
	for num, _ := range result.Quotes {
		tablename := result.Quotes[num].Code
		stmtIns, _ = db.Prepare(fmt.Sprintf("INSERT INTO %s (Time,Open,Bid,Ask,High,Low) VALUES (?,?,?,?,?,?)", tablename))
		_, _ = stmtIns.Exec(result.time, result.Quotes[num].Open, result.Quotes[num].Bid, result.Quotes[num].Ask, result.Quotes[num].High, result.Quotes[num].Low)
	}
	defer stmtIns.Close()
}

func for_db() {
	for {
		InDataBase()
		time.Sleep(10 * time.Minute)
	}
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

	go for_db()

	r := mux.NewRouter()
	r.HandleFunc("/", RootHandler)
	r.HandleFunc("/top", TopHandler)
	/* ルーティング設定 */
	r.HandleFunc("/detail/{name}", DataHandler)
	r.HandleFunc("/auth", SecretHandler)
	r.HandleFunc("/settings", SettingsHandler)
	r.HandleFunc("/redirect", RedirectHandler)
	log.Fatal(http.ListenAndServe(":8080", r))
}
