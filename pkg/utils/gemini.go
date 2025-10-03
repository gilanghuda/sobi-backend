package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

// QueryGemini sends the given prompt to a configured Gemini-compatible API
// endpoint and returns a textual reply. Configure endpoint and key with
// GEMINI_API_URL and GEMINI_API_KEY environment variables.
func QueryGemini(prompt string) (string, error) {
	url := os.Getenv("GEMINI_API_URL")
	key := os.Getenv("GEMINI_API_KEY")
	if url == "" || key == "" {
		return "", errors.New("gemini not configured")
	}

	reqBody := map[string]interface{}{"prompt": prompt}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

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

	// Try common keys for text output
	if out, ok := parsed["output"].(string); ok && out != "" {
		return out, nil
	}
	if out, ok := parsed["response"].(string); ok && out != "" {
		return out, nil
	}
	if out, ok := parsed["text"].(string); ok && out != "" {
		return out, nil
	}
	if choices, ok := parsed["choices"].([]interface{}); ok && len(choices) > 0 {
		first := choices[0]
		switch f := first.(type) {
		case map[string]interface{}:
			if t, ok := f["text"].(string); ok && t != "" {
				return t, nil
			}
			if m, ok := f["message"].(map[string]interface{}); ok {
				if t2, ok := m["content"].(string); ok && t2 != "" {
					return t2, nil
				}
			}
		}
	}

	// fallback: marshal the whole response
	b2, _ := json.Marshal(parsed)
	return string(b2), nil
}
