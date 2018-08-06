package smugmug

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
		return fmt.Errorf("Server returned %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("Body: %s\n", string(body[:]))
	err = json.Unmarshal(body, &target)
	if err != nil {
		return err
	}
	return nil
}
