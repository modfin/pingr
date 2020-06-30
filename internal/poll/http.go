package poll

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

func HTTP(hostname string, method string, to time.Duration, payload []byte, expRes []byte) (rt time.Duration, err error) {
	client := http.Client{Timeout: to}
	start := time.Now()
	req, err := http.NewRequest(method, hostname, bytes.NewBuffer(payload))
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	rt = time.Since(start)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if expRes != nil && !bytes.Equal(body, expRes) {
		err = errors.New("HTTP request response is not matching the expected result")
		return
	}

	return
}