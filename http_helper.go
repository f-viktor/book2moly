package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func updateCookieJar(current *[]*http.Cookie, new *[]*http.Cookie) {
	for _, newCookie := range *new {

		//update cookie value if already exists
		update := false
		for _, currentCookie := range *current {
			if newCookie.Name == currentCookie.Name {
				currentCookie.Value = newCookie.Value
				update = true
			}
		}

		//append cookie if it doesn't
		if !update {
			*current = append(*current, newCookie)
		}
	}
}

func performHTTPRequest(method string, reqURL string, body []byte, cookies []*http.Cookie) ([]byte, []*http.Cookie) {
	req, _ := http.NewRequest(method, reqURL, bytes.NewBuffer(body))

	//parse this from url
	req.Host = strings.Split(reqURL, "/")[2]
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36")
	req.Header.Set("Content-Type", "text/html; charset=utf-8")
	//req.Header.Set("Origin", "https://moly.hu")

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// add cookies
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	tr := &http.Transport{}
	if GlobalConfig.HttpProxy != "" {
		/*Debug feature to Turn off cert validation */
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		// this is for debug proxying
		proxy, _ := url.Parse(GlobalConfig.HttpProxy)
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyURL(proxy),
		}
	}

	//for avoiding infinite redirect loops
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[!] HTTP request failed to " + method + " " + reqURL)
		fmt.Printf(err.Error())
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 302 {
		//returned with <status code>
		fmt.Printf("[!] HTTP request returned with " + strconv.Itoa(resp.StatusCode) + " : " + method + " " + reqURL)

	}

	return []byte(respBody), resp.Cookies()

}

// There is no real reason for having this in a separate function apart from the fact that I find it disgusting
func performMultiPartForm(reqURL string, values map[string]io.Reader, cookies []*http.Cookie) ([]byte, []*http.Cookie) {
	var err error
	var bodyBuffer bytes.Buffer
	bodyWriter := multipart.NewWriter(&bodyBuffer)

	for fieldName, fieldReader := range values {
		var fieldWriter io.Writer

		if fieldValue, ok := fieldReader.(io.Closer); ok {
			defer fieldValue.Close()
		}

		// Add an image file
		if file, ok := fieldReader.(*os.File); ok {
			if fieldWriter, err = bodyWriter.CreateFormFile(fieldName, file.Name()); err != nil {
				panic("[!] File could not be read")
			}
		} else {
			// Add other fields
			if fieldWriter, err = bodyWriter.CreateFormField(fieldName); err != nil {
				panic("[!] Field could not be read")
			}
		}

		if _, err = io.Copy(fieldWriter, fieldReader); err != nil {
			panic(err)
		}
	}

	// If you don't close it, your request will be missing the terminating boundary.
	bodyWriter.Close()

	req, _ := http.NewRequest("POST", reqURL, &bodyBuffer)
	//for multipart form
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	//parse this from url
	req.Host = strings.Split(reqURL, "/")[2]
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36")
	req.Header.Set("Content-Type", "text/html; charset=utf-8")
	//req.Header.Set("Origin", "https://moly.hu")

	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	// add cookies
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	tr := &http.Transport{}
	if GlobalConfig.HttpProxy != "" {
		/*Debug feature to Turn off cert validation */
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		// this is for debug proxying
		proxy, _ := url.Parse(GlobalConfig.HttpProxy)
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyURL(proxy),
		}
	}

	//for avoiding infinite redirect loops
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[!] HTTP request failed to " + "POST" + " " + reqURL)
		fmt.Printf(err.Error())
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 302 {
		//returned with <status code>
		fmt.Printf("[!] HTTP request returned with " + strconv.Itoa(resp.StatusCode) + " : " + "POST" + " " + reqURL)

	}

	return []byte(respBody), resp.Cookies()

}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}
