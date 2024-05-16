package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type ExchangeRate struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type Response struct {
	value ExchangeRate
	err   error
}

func Server() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /cotacao", getCotacao)
	http.ListenAndServe(":8080", mux)
}

func getCotacao(w http.ResponseWriter, r *http.Request) {
	apiTimeoutMili := 200
	apiCtx, cancel := context.WithTimeout(context.Background(), time.Duration(apiTimeoutMili)*time.Millisecond)
	defer cancel()

	respApiCh := make(chan Response)

	go func() {
		val, err := getLatestExchangeRateUSDBRL(apiCtx)
		respApiCh <- Response{
			value: val,
			err:   err,
		}
	}()

	for {
		select {
		case <-apiCtx.Done():
			w.WriteHeader(http.StatusRequestTimeout)
			return
		case resp := <-respApiCh:
			if resp.err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp.value.USDBRL.Bid)
		}
	}
}

func getLatestExchangeRateUSDBRL(ctx context.Context) (ExchangeRate, error) {
	apiURL := "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return ExchangeRate{}, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ExchangeRate{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ExchangeRate{}, err
	}
	var ex ExchangeRate
	err = json.Unmarshal(body, &ex)
	if err != nil {
		return ExchangeRate{}, err
	}
	return ex, nil
}
