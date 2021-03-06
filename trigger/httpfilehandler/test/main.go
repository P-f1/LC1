package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"strings"
)

func main() {
	Upload2()
}

func mainx() {

	var client *http.Client
	var remoteURL string
	{
		//setup a mocked http client.
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := httputil.DumpRequest(r, true)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s", b)
		}))
		defer ts.Close()
		client = ts.Client()
		remoteURL = ts.URL
	}

	//prepare the reader instances to encode
	values := map[string]io.Reader{
		"file":  mustOpen("main.go"), // lets assume its this file
		"other": strings.NewReader("hello world!"),
	}
	err := Upload(client, remoteURL, values)
	if err != nil {
		panic(err)
	}
}

func Upload(client *http.Client, url string, values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}

func Upload2() {
	file, err := os.Open("./SteveNY-Tibco_iris-classifier-pickle.tar.gz")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	res, err := http.Post("http://127.0.0.1:5408/mops/projectmgr/pushProject/SteveNY-Tibco_iris-classifier-pickle.tar.gz", "binary/octet-stream", file)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	message, _ := ioutil.ReadAll(res.Body)
	fmt.Printf(string(message))
}
