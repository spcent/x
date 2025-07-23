package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func VerifyTokenWithGateway(token string, gatewayURL string) (bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/verify-token", gatewayURL), nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("x-token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("invalid token")
	}

	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return false, err
	}

	status, ok := response["status"].(bool)
	return ok && status, nil
}
