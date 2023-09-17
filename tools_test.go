package toolkit

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Errorf("RandomString failed, length of string is %d, expected 10", len(s))
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	rename        bool
	errorExpected bool
}{
	{
		name:          "Allowed no rename",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		rename:        false,
		errorExpected: false,
	},
	{
		name:          "Allowed rename",
		allowedTypes:  []string{"image/jpeg", "image/png"},
		rename:        true,
		errorExpected: false,
	},
	{
		name:          "not allowed",
		allowedTypes:  []string{"image/jpeg"},
		rename:        false,
		errorExpected: true,
	},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {

		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer writer.Close()
			defer wg.Done()

			// create the form data field
			part, err := writer.CreateFormFile("file", "./testdata/img.png")
			if err != nil {
				t.Error(err)
			}
			f, err := os.Open("./testdata/img.png")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()
			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("error decoding image", err)
			}
			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}
		}()

		// read from the pipe which receives data
		request := httptest.NewRequest("POST", "/upload", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads", e.rename)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}
		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {

				t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
			}
			_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}
		if !e.errorExpected && err != nil {
			t.Errorf("%s: expected no error received", e.name)
		}

		wg.Wait()
	}
}

func TestTools_UploadOneFile(t *testing.T) {
	for _, e := range uploadTests {

		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)

		go func() {
			defer writer.Close()

			// create the form data field
			//writer.WriteField("foo", "bar")
			part, err := writer.CreateFormFile("file", "./testdata/img.png")
			if err != nil {
				t.Error(err)
			}
			f, err := os.Open("./testdata/img.png")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()
			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("error decoding image", err)
			}
			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}
		}()

		// read from the pipe which receives data
		request := httptest.NewRequest("POST", "/upload", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools

		uploadedFiles, err := testTools.UploadOneFile(request, "./testdata/uploads", true)
		if err != nil {
			t.Error(err)
		}
		if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName)); os.IsNotExist(err) {

			t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
		}
		_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName))
		if !e.errorExpected && err != nil {
			t.Errorf("%s: expected no error received", e.name)
		}
	}
}

func TestTools_CreateDirIfNotExist(t *testing.T) {
	var testTools Tools

	err := testTools.CreateDirIfNotExist("./testdata/myDir")
	if err != nil {
		t.Error(err)
	}

	err = testTools.CreateDirIfNotExist("./testdata/myDir")
	if err != nil {
		t.Error(err)
	}
	_ = os.Remove("./testdata/myDir")

}

var slugTests = []struct {
	name          string
	s             string
	expected      string
	errorExpected bool
}{
	{name: "Valid string", s: "now is the time", expected: "now-is-the-time", errorExpected: false},
	{name: "Empty string", s: "", expected: "", errorExpected: true},
	{name: "Complex string", s: "Now is the time for all GOOD men! + fish & such &^", expected: "now-is-the-time-for-all-good-men-fish-such", errorExpected: true},
	{name: "Japanese string", s: "こんにちは世界", expected: "", errorExpected: true},
	{name: "Japanese string and roman characters", s: "hello world こんにちは世界", expected: "hello-world", errorExpected: false},
}

func TestTools_Slugify(t *testing.T) {
	var testTool Tools
	for _, e := range slugTests {
		slug, err := testTool.Slugify(e.s)
		if err != nil && !e.errorExpected {
			t.Errorf("%s: expected received when none expected: %s", e.name, err.Error())
		}

		if !e.errorExpected && slug != e.expected {
			t.Errorf("%s: expected %s, received %s", e.name, e.expected, slug)
		}
	}
}
