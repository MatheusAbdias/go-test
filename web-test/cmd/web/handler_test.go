package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"testing"

	"webapp/pkg/data"
)

func Test_application_handlers(t *testing.T) {
	var testCases = []struct {
		name                    string
		url                     string
		expectedStatusCode      int
		expectedURL             string
		expectedFirstStatusCode int
	}{
		{"home", "/", http.StatusOK, "/", http.StatusOK},
		{"404", "/fish", http.StatusNotFound, "/fish", http.StatusNotFound},
		{"profile", "/user/profile", http.StatusOK, "/", http.StatusTemporaryRedirect},
	}

	routes := app.routes()

	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, testCase := range testCases {
		response, err := ts.Client().Get(ts.URL + testCase.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if response.StatusCode != testCase.expectedStatusCode {
			t.Errorf(
				"for %s:expected status code %d but got %d",
				testCase.name,
				testCase.expectedStatusCode,
				response.StatusCode,
			)
		}

		if response.Request.URL.Path != testCase.expectedURL {
			t.Errorf(
				"%s: expected final url of %s but got %s",
				testCase.name,
				testCase.expectedURL,
				response.Request.URL.Path,
			)
		}

		response2, _ := client.Get(ts.URL + testCase.url)

		if response2.StatusCode != testCase.expectedFirstStatusCode {
			t.Errorf(
				"%s: expected first returned status code to be %d but got %d",
				testCase.name,
				testCase.expectedFirstStatusCode,
				response2.StatusCode,
			)
		}
	}
}

func TestAppHome(t *testing.T) {
	var testCases = []struct {
		name         string
		putInSession string
		expectedHTML string
	}{
		{"First visit", "", "<small>From Session:"},
		{"Second visit", "hello, world!", "<small>From Session: hello, world!"},
	}

	for _, testCase := range testCases {
		request, _ := http.NewRequest("GET", "/", nil)

		request = addContextAndSessionToRequest(request, app)
		_ = app.Session.Destroy(request.Context())

		if testCase.putInSession != "" {
			app.Session.Put(request.Context(), "test", testCase.putInSession)
		}

		recorder := httptest.NewRecorder()

		handler := http.HandlerFunc(app.Home)

		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf(
				"TestAppHome returned wrong status code; expected %d but got: %d",
				http.StatusOK,
				recorder.Code,
			)
		}

		body, _ := io.ReadAll(recorder.Body)
		if !strings.Contains(string(body), testCase.expectedHTML) {
			t.Errorf("%s: did not found %s in response body", testCase.name, testCase.expectedHTML)
		}
	}
}

func TestApp_render_WithBadTemplate(t *testing.T) {
	pathToTemplates = "./testdata/"

	request, _ := http.NewRequest("GET", "/", nil)
	request = addContextAndSessionToRequest(request, app)
	render := httptest.NewRecorder()

	err := app.render(render, request, "bad.page.gohtml", &TemplateData{})
	if err == nil {
		t.Error("expected error from bad template, but did not get one")
	}

	pathToTemplates = "./templates/"

}

func getCtx(request *http.Request) context.Context {
	return context.WithValue(request.Context(), contextUserKey, "unknown")
}

func addContextAndSessionToRequest(request *http.Request, app application) *http.Request {
	request = request.WithContext(getCtx(request))

	ctx, _ := app.Session.Load(request.Context(), request.Header.Get("X-Session"))

	return request.WithContext(ctx)
}

func Test_app_Login(t *testing.T) {
	var testCases = []struct {
		name               string
		postedData         url.Values
		expectedStatusCode int
		expectedLoc        string
	}{
		{
			name: "valid login",
			postedData: url.Values{
				"email":    {"admin@example.com"},
				"password": {"secret"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "user/profile",
		},
		{
			name: "missing form data",
			postedData: url.Values{
				"email":    {""},
				"password": {""},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/",
		},
		{
			name: "bad credentials",
			postedData: url.Values{
				"email":    {"admin@example.com"},
				"password": {"wrong password"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/",
		},
		{
			name: "use doest exist",
			postedData: url.Values{
				"email":    {"admin3@example.com"},
				"password": {"secret"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/",
		},
	}

	for _, testCase := range testCases {
		request, _ := http.NewRequest(
			"POST",
			"login",
			strings.NewReader(testCase.postedData.Encode()),
		)
		request = addContextAndSessionToRequest(request, app)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()
		handler := http.HandlerFunc(app.Login)
		handler.ServeHTTP(recorder, request)

		if recorder.Code != testCase.expectedStatusCode {
			t.Errorf(
				"%s: return wrong status code; expected %d, but got %d",
				testCase.name,
				testCase.expectedStatusCode,
				recorder.Code,
			)
		}

		actualLoc, err := recorder.Result().Location()
		if err == nil {

			if actualLoc.String() != testCase.expectedLoc {
				t.Errorf(
					"%s: expected location %s but got %s",
					testCase.name,
					testCase.expectedLoc,
					actualLoc.String(),
				)
			}
		} else {
			t.Errorf("%s: no location header set", testCase.name)
		}
	}
}

func Test_App_UploadFile(t *testing.T) {
	pr, pw := io.Pipe()

	writer := multipart.NewWriter(pw)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go simulatePNGUpload("./testdata/img.png", writer, t, wg)

	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	uploadedFiles, err := app.UploadFiles(request, "./testdata/uploads/")
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].OriginalFileName)); os.IsNotExist(
		err,
	) {
		t.Errorf("expected file to exists: %s", err.Error())
	}

	_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].OriginalFileName))

	wg.Wait()
}

func simulatePNGUpload(
	fileToUpload string,
	writer *multipart.Writer,
	t *testing.T,
	wg *sync.WaitGroup,
) {
	defer writer.Close()
	defer wg.Done()

	part, err := writer.CreateFormFile("file", path.Base(fileToUpload))
	if err != nil {
		t.Error(err)
	}

	f, err := os.Open(fileToUpload)
	if err != nil {
		t.Error(err)
	}

	defer f.Close()

	img, _, err := image.Decode(f)

	if err != nil {
		t.Error("error decoding image:", err)
	}

	err = png.Encode(part, img)
	if err != nil {
		t.Error(err)
	}
}

func Test_app_UploadProfilePic(t *testing.T) {
	uploadPath = "./testdata/uploads"
	filePath := "./testdata/img.png"
	fieldName := "file"

	body := new(bytes.Buffer)

	mw := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}

	w, err := mw.CreateFormFile(fieldName, filePath)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.Copy(w, file); err != nil {
		t.Fatal(err)
	}

	mw.Close()

	request := httptest.NewRequest(http.MethodPost, "/upload", body)
	request = addContextAndSessionToRequest(request, app)

	app.Session.Put(request.Context(), "user", data.User{ID: 1})
	request.Header.Add("Content-Type", mw.FormDataContentType())

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(app.UploadProfilePic)

	handler.ServeHTTP(rr, request)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code")
	}

	_ = os.Remove("./testdata/uploads/img.png")
}
