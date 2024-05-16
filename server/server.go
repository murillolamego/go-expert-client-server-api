package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRate struct {
	USDBRL struct {
		code       string
		codein     string
		name       string
		high       string
		low        string
		varBid     string
		pctChange  string
		Bid        string `json:"bid"`
		ask        string
		timestamp  string
		createDate string
	} `json:"USDBRL"`
}

type ApiResponse struct {
	value ExchangeRate
	err   error
}

type DbResponse struct {
	err error
}

func main() {

	const create string = `
	CREATE TABLE IF NOT EXISTS usdbrl (
	id INTEGER NOT NULL PRIMARY KEY,
	code TEXT,
	codein TEXT,
	name TEXT,
	high TEXT,
	low TEXT,
	var_bid TEXT,
	pct_change TEXT,
	bid TEXT,
	ask TEXT,
	timestamp TEXT,
	create_date TEXT
	);`

	db, err := sql.Open("sqlite3", "usdbrl.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(create)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /cotacao", cotacaoHandler)
	http.ListenAndServe(":8080", mux)
}

func cotacaoHandler(w http.ResponseWriter, r *http.Request) {
	apiTimeoutMili := 200
	apiCtx, apiCancel := context.WithTimeout(context.Background(), time.Duration(apiTimeoutMili)*time.Millisecond)
	defer apiCancel()
	respApiCh := make(chan ApiResponse)

	dbTimeoutMili := 10
	dbCtx, dbCancel := context.WithTimeout(context.Background(), time.Duration(dbTimeoutMili)*time.Millisecond)
	defer dbCancel()
	respDbCh := make(chan DbResponse)

	go func() {
		val, err := getLatestExchangeRateUSDBRL(apiCtx)
		respApiCh <- ApiResponse{
			value: val,
			err:   err,
		}
	}()

	for {
		select {
		case <-apiCtx.Done():
			println("Err: Fetching data from external API took too long")
			w.WriteHeader(http.StatusRequestTimeout)
			return
		case apiResp := <-respApiCh:
			if apiResp.err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			go func() {
				err := persistLatestExchangeRateUSDBRL(dbCtx, apiResp.value)
				respDbCh <- DbResponse{
					err: err,
				}
			}()

			for {
				select {
				case <-dbCtx.Done():
					println("Err: Writing to database took too long")
					w.WriteHeader(http.StatusRequestTimeout)
					return
				case dbResp := <-respDbCh:
					if dbResp.err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(apiResp.value.USDBRL)
					return
				}
			}
		}
	}
}

func getLatestExchangeRateUSDBRL(ctx context.Context) (ExchangeRate, error) {
	apiURL := "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	var ex ExchangeRate
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return ex, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ex, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ex, err
	}
	err = json.Unmarshal(body, &ex)
	if err != nil {
		return ex, err
	}
	return ex, nil
}

func persistLatestExchangeRateUSDBRL(ctx context.Context, ex ExchangeRate) error {
	c := ex.USDBRL

	db, err := sql.Open("sqlite3", "usdbrl.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "INSERT INTO usdbrl VALUES(NULL,?,?,?,?,?,?,?,?,?,?,?);",
		c.code,
		c.codein,
		c.name,
		c.high,
		c.low,
		c.varBid,
		c.pctChange,
		c.Bid,
		c.ask,
		c.timestamp,
		c.createDate)
	if err != nil {
		println(err.Error())
		return err
	}

	return nil
}
