package main

import (
	"net/http"
	"strings"
	"testing"

	chi "github.com/go-chi/chi/v5"
)

func Test_application_routes(t *testing.T) {
	var testCases = []struct {
		route  string
		method string
	}{
		{"/", "GET"},
		{"/login", "POST"},
		{"/user/profile", "GET"},
		{"/static/*", "GET"},
	}

	mux := app.routes()

	chiRoutes := mux.(chi.Routes)

	for _, testCase := range testCases {
		if !routeExists(testCase.route, testCase.method, chiRoutes) {
			t.Errorf("route %s %s does not exist", testCase.method, testCase.route)
		}
	}
}

func routeExists(testRoute, testMethod string, chiRoutes chi.Routes) bool {
	found := false

	_ = chi.Walk(
		chiRoutes,
		func(method string, route string, handler http.Handler, middleware ...func(http.Handler) http.Handler) error {
			if strings.EqualFold(method, testMethod) && strings.EqualFold(route, testRoute) {
				found = true
			}
			return nil
		},
	)
	return found
}
