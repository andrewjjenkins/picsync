package googlephotos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// GetUnmarshalJSON gets a JSON response from url and unmarshals into target
func GetUnmarshalJSON(c *http.Client, url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("accept", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("Body: %s\n", body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &target)
	if err != nil {
		return err
	}
	return nil
}

func PostUnmarshalJSON(c *http.Client, url string, reqBody string, target interface{}) error {
	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("Body: %s\n", body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &target)
	if err != nil {
		return err
	}
	return nil
}
