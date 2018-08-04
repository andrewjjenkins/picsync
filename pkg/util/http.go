package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetUnmarshalJSON gets a JSON response from url and unmarshals into target
func GetUnmarshalJSON(c *http.Client, url string, target interface{}) error {
	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Server returned %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, target)
	if err != nil {
		return err
	}
	return nil
}
