// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/28 by Vincent Landgraf

package runtime

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUnmarshalAccept(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", nil)

	ok := Unmarshal(rec, req, nil)
	if ok {
		t.Error("Un-marshalling should fail")
	}

	resp := rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotAcceptable {
		t.Errorf("Expected status code %d got: %d", http.StatusNotAcceptable, resp.StatusCode)
	}
}
func TestUnmarshalContentType(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Accept", JSONAPIContentType)

	ok := Unmarshal(rec, req, nil)
	if ok {
		t.Error("Un-marshalling should fail")
	}

	resp := rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("Expected status code %d got: %d", http.StatusUnsupportedMediaType, resp.StatusCode)
	}
}
func TestUnmarshalContent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"data": 1}`))
	req.Header.Set("Accept", JSONAPIContentType)
	req.Header.Set("Content-Type", JSONAPIContentType)

	type Article struct {
		ID    string `jsonapi:"id,articles" valid:"optional,uuid"`
		Title string `jsonapi:"title" valid:"required"`
	}

	var article Article
	ok := Unmarshal(rec, req, &article)
	if ok {
		t.Error("Un-marshalling should fail")
	}

	resp := rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status code %d got: %d", http.StatusUnprocessableEntity, resp.StatusCode)
	}
}

func TestUnmarshalArticle(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"data":{
		"type": "articles",
		"id": "cb855aff-f03c-4307-9a22-ab5fcc6b6d7c",
		"attributes": {
			"title": "This is my first blog"
		}
	}}`))
	req.Header.Set("Accept", JSONAPIContentType)
	req.Header.Set("Content-Type", JSONAPIContentType)

	type Article struct {
		ID    string `jsonapi:"primary,articles" valid:"optional,uuid"`
		Title string `jsonapi:"attr,title" valid:"required"`
	}

	var article Article
	ok := Unmarshal(rec, req, &article)

	if !ok {
		t.Error("Un-marshalling should have been ok")
	}

	resp := rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d got: %d", http.StatusOK, resp.StatusCode)
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Error(string(b[:]))
	}

	uuid := "cb855aff-f03c-4307-9a22-ab5fcc6b6d7c"
	if article.ID != uuid {
		t.Errorf("article.ID expected %q got: %q", uuid, article.ID)
	}
	if article.Title != "This is my first blog" {
		t.Errorf("article.ID expected \"This is my first blog\" got: %q", article.Title)
	}
}
func TestMarshalArticle(t *testing.T) {
	rec := httptest.NewRecorder()

	type Article struct {
		ID    string `jsonapi:"primary,articles" valid:"optional,uuid"`
		Title string `jsonapi:"attr,title" valid:"required"`
	}

	article := Article{
		ID:    "cb855aff-f03c-4307-9a22-ab5fcc6b6d7c",
		Title: "This is my first blog",
	}
	Marshal(rec, &article, http.StatusOK)

	resp := rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d got: %d", http.StatusOK, resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expectation := `{"data":{"type":"articles","id":"cb855aff-f03c-4307-9a22-ab5fcc6b6d7c","attributes":{"title":"This is my first blog"}}}` + "\n"

	if expectation != string(b[:]) {
		t.Errorf("Expected %q got: %q", expectation, string(b[:]))
	}
}
