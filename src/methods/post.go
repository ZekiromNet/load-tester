package methods

import (
	"net/http"
	"strings"
)

func DoPost(client *http.Client, url string) (int, error) {
	resp, err := client.Post(url, "application/x-www-form-urlencoded", strings.NewReader(""))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}
