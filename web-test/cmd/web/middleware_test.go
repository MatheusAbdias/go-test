package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"webapp/pkg/data"
)

func Test_application_addIPToContext(t *testing.T) {
	testCases := []struct {
		headerName  string
		headerValue string
		addr        string
		emptyAddr   bool
	}{
		{"", "", "", false},
		{"", "", "", true},
		{"X-Forwarded-For", "192.3.2.1", "", false},
		{"", "", "hello:world", false},
	}
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Context().Value(contextUserKey)
		if value == nil {
			t.Error(contextUserKey, "not present")
		}

		ip, ok := value.(string)
		if !ok {
			t.Error("not string")
		}
		t.Log(ip)
	})

	for _, testCase := range testCases {
		handleToTest := app.addIPToContext(nextHandler)

		request := httptest.NewRequest("GET", "http://testing", nil)

		if testCase.emptyAddr {
			request.RemoteAddr = ""
		}

		if len(testCase.headerName) > 0 {
			request.Header.Add(testCase.headerName, testCase.headerValue)
		}

		if len(testCase.addr) > 0 {
			request.RemoteAddr = testCase.addr
		}

		handleToTest.ServeHTTP(httptest.NewRecorder(), request)
	}
}

func Test_application_ipFromContext(t *testing.T) {
	testCases := []struct {
		headerValue string
	}{
		{"192.3.2.1"},
		{"fool"},
	}
	var ctx context.Context

	for _, testCase := range testCases {
		ctx = context.WithValue(context.Background(), contextUserKey, testCase.headerValue)
		ip := app.ipFromContext(ctx)

		if !strings.EqualFold(ip, testCase.headerValue) {
			t.Errorf("Expected:%s, but got:%s", testCase.headerValue, ip)
		}
	}

}

func Test_app_auth(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	testCases := []struct {
		name   string
		isAuth bool
	}{
		{"logged in", true},
		{"not legged in", false},
	}
	for _, testCase := range testCases {
		handlerToTest := app.auth(nextHandler)
		request := httptest.NewRequest("GET", "http://testing", nil)
		request = addContextAndSessionToRequest(request, app)
		if testCase.isAuth {
			app.Session.Put(request.Context(), "user", data.User{ID: 1})
		}

		recorder := httptest.NewRecorder()
		handlerToTest.ServeHTTP(recorder, request)

		if testCase.isAuth && recorder.Code != http.StatusOK {
			t.Errorf("%s: expected status code 200 but got %d", testCase.name, recorder.Code)
		}

		if !testCase.isAuth && recorder.Code != http.StatusTemporaryRedirect {
			t.Errorf(
				"%s: expected status code %d but got %d",
				testCase.name,
				http.StatusTemporaryRedirect,
				recorder.Code,
			)
		}
	}

}
