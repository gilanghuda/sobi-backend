package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

func CreateMidtransTransaction(orderID string, amount int64) (string, error) {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	if serverKey == "" {
		return "", errors.New("midtrans server key not set")
	}

	endpoint := "https://app.sandbox.midtrans.com/snap/v1/transactions"
	payload := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":     orderID,
			"gross_amount": amount,
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	auth := base64.StdEncoding.EncodeToString([]byte(serverKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return "", errors.New("midtrans returned error status")
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return "", err
	}

	if ru, ok := resp["redirect_url"].(string); ok && ru != "" {
		return ru, nil
	}
	if token, ok := resp["token"].(string); ok && token != "" {
		return "https://app.sandbox.midtrans.com/snap/v2/vtweb/" + token, nil
	}

	return "", errors.New("no redirect_url returned from midtrans")
}
