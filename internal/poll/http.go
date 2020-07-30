package poll

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func HTTP(hostname string, method string, to time.Duration, reqHeaders map[string]string, reqBody string, resStatus int, resHeaders map[string]string, resBody string) (time.Duration, error) {
	client := http.Client{Timeout: to}
	start := time.Now()

	req, err := http.NewRequest(method, hostname, bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return time.Since(start), err
	}

	// Set headers
	for k, v := range reqHeaders {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return time.Since(start), err
	}
	defer resp.Body.Close()
	rt := time.Since(start)

	if resStatus != 0 && resp.StatusCode != resStatus {
		return rt, fmt.Errorf("response statuscode is not matching the expected value, got: %d, expected: %d", resp.StatusCode, resStatus)
	}

	// Header check
	for k, v := range resHeaders {
		if resp.Header.Get(k) != v {
			return rt, fmt.Errorf("response header is not matching expected header, key: %s, got: %s, expected: %s", k, resp.Header.Get(k), v)
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rt, err
	}

	// Body check
	if resBody != "" && string(body) != resBody {
		err = fmt.Errorf("response body is not matching expected body, got: %s, expected: %s", string(body), resBody)
		return rt, err
	}

	return rt, err
}
