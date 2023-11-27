package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Has(t *testing.T) {
	form := NewForm(nil)

	has := form.Has("fool")

	if has {
		t.Error("form show has field when it should not")
	}

	postedData := url.Values{}

	postedData.Add("a", "a")
	form = NewForm(postedData)

	has = form.Has("a")
	if !has {
		t.Error("form show no has field when it should")
	}
}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/fool", nil)
	form := NewForm(r.PostForm)

	form.Required("a", "b", "c")

	if form.Valid() {
		t.Error("form show is valid when required fields are missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	r, _ = http.NewRequest("POST", "/fool", nil)
	r.PostForm = postedData

	form = NewForm(r.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("form show is not valid when all required field are fielded")
	}
}

func TestForm_Check(t *testing.T) {
	form := NewForm(nil)

	form.Check(false, "password", "password is required")

	if form.Valid() {
		t.Error("valid returns false,and it should be true when calling Check")
	}
}

func TestForm_ErrorGet(t *testing.T) {
	form := NewForm(nil)
	form.Check(false, "password", "password is required")
	s := form.Errors.Get("password")

	if len(s) == 0 {
		t.Error("should have an error returned from get, but do not")
	}

	s = form.Errors.Get("email")

	if len(s) != 0 {
		t.Error("should not have an error, but got one")
	}
}
