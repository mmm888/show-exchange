package db_in

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"time"
)

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

