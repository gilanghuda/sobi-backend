package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

func QueryGemini(prompt string) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-flash-latest:generateContent"
	key := os.Getenv("GEMINI_API_KEY")
	if url == "" || key == "" {
		return "", errors.New("gemini not configured")
	}

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	b, _ := json.Marshal(reqBody)

	fullURL := url + "?key=" + key
	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", errors.New("gemini request failed: " + string(bodyBytes))
	}

	var parsed map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}

	if cts, ok := parsed["candidates"].([]interface{}); ok && len(cts) > 0 {
		first := cts[0].(map[string]interface{})
		if content, ok := first["content"].(map[string]interface{}); ok {
			if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
				if text, ok := parts[0].(map[string]interface{})["text"].(string); ok {
					return text, nil
				}
			}
		}
	}

	b2, _ := json.Marshal(parsed)
	return string(b2), nil
}
