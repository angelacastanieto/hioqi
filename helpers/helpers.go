package helpers

import (
	"fmt"
	"net/http"
)

const (
	HeaderAuthorization = "Authorization"
)

func Get(client *http.Client, endpoint, token string) (*http.Response, error) {
	var resp *http.Response

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return resp, err
	}

	req.Header.Add(HeaderAuthorization, fmt.Sprintf("Bearer %s", token))

	resp, err = client.Do(req)

	return resp, err
}
