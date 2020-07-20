package poll

import (
	"bytes"
	"errors"
	"github.com/jmoiron/sqlx/types"
	"io/ioutil"
	"net/http"
	"time"
)

func HTTP(hostname string, method string, to time.Duration, payload types.JSONText, expRes types.JSONText) (time.Duration, error) {
	client := http.Client{Timeout: to}
	start := time.Now()
	req, err := http.NewRequest(method, hostname, bytes.NewBuffer(payload))
	if err != nil {
		return time.Since(start), err
	}
	resp, err := client.Do(req)
	if err != nil {
		return time.Since(start), err
	}
	defer resp.Body.Close()
	rt := time.Since(start)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rt, err
	}

	if expRes != nil && !bytes.Equal(body, expRes) {
		err = errors.New("HTTP request response is not matching the expected result")
		return rt, err
	}

	return rt, err
}