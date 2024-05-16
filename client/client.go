package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

type ServerResponse struct {
	value Cotacao
	err   error
}

func main() {
	serverTimeoutMili := 300
	serverCtx, serverCancel := context.WithTimeout(context.Background(), time.Duration(serverTimeoutMili)*time.Millisecond)
	defer serverCancel()
	respServerCh := make(chan ServerResponse)

	go func() {
		val, err := requestCotacao(serverCtx)
		respServerCh <- ServerResponse{
			value: val,
			err:   err,
		}
	}()

	for {
		select {
		case <-serverCtx.Done():
			println("Err: Fetching data from Server took too long")
			return
		case serverResp := <-respServerCh:
			if serverResp.err != nil {
				panic(serverResp.err)
			}
			persistCotacao(serverResp.value.Bid)
			return
		}
	}
}

func requestCotacao(ctx context.Context) (Cotacao, error) {
	serverURL := "http://localhost:8080/cotacao"

	var r Cotacao
	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if err != nil {
		return r, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return r, err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 408:
		panic("Server took too long to respond")
	case 500:
		panic("Server error")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return r, err
	}
	if len(body) == 0 {
		panic("Empty response from server")
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

func persistCotacao(cotacao string) {

	txt := fmt.Sprintf("Dólar: {%s}", cotacao)

	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.WriteString(txt)
	if err != nil {
		panic(err)
	}
}
