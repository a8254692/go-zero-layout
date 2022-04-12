package help

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

var tr = &http.Transport{ //解决x509: certificate signed by unknown authority
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var httpClient *http.Client = &http.Client{
	Timeout:   time.Duration(3 * time.Second),
	Transport: tr,
}

func GetUtilHttpClient() *http.Client {
	return httpClient
}

func OnPostHttp(uri string, body []byte, header map[string]string) (int, []byte) {
	req, err := http.NewRequest("POST", uri, bytes.NewReader(body))
	if err != nil {
		logx.Error("OnPostHttp.post %s, error = %v", uri, err)
		return http.StatusInternalServerError, nil
	}

	if header != nil {
		for k, v := range header {
			req.Header.Add(k, v)
		}
	}

	logx.Info("OnPostHttp.uri = %s, done", uri)
	resp, err := httpClient.Do(req)
	if err != nil {
		logx.Error("OnPostHttp.post %s, error = %v", uri, err)
		return http.StatusInternalServerError, nil
	}

	data, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logx.Error("OnPostHttp.post %s, status = %d", uri, resp.StatusCode)
		return resp.StatusCode, data
	}

	return http.StatusOK, data
}

func OnGetHttp(uri string, header map[string]string) (int, []byte) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		logx.Error("OnGetHttp.get %s, error = %v", uri, err)
		return http.StatusInternalServerError, nil
	}

	if header != nil {
		for k, v := range header {
			req.Header.Add(k, v)
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		logx.Error("OnGetHttp.get %s, error = %v", uri, err)
		return http.StatusInternalServerError, nil
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logx.Error("OnGetHttp.get %s, status = %d", uri, resp.StatusCode)
		return resp.StatusCode, []byte(resp.Status)
	}

	logx.Info("OnGetHttp.uri = %s, done", uri)
	data, _ := ioutil.ReadAll(resp.Body)
	return http.StatusOK, data
}
