// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package runtime

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalAccept(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	ok := Unmarshal(rec, req, nil)
	if ok {
		t.Error("Un-marshalling should fail")
	}

	resp := rec.Result()

	defer func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	}()

	if resp.StatusCode != http.StatusNotAcceptable {
		t.Errorf("Expected status code %d got: %d", http.StatusNotAcceptable, resp.StatusCode)
	}
}

func TestUnmarshalContentType(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Accept", JSONAPIContentType)

	ok := Unmarshal(rec, req, nil)
	if ok {
		t.Error("Un-marshalling should fail")
	}

	resp := rec.Result()

	defer func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	}()

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("Expected status code %d got: %d", http.StatusUnsupportedMediaType, resp.StatusCode)
	}
}

func TestUnmarshalContent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"data": 1}`))
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

	defer func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	}()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status code %d got: %d", http.StatusUnprocessableEntity, resp.StatusCode)
	}
}

func TestUnmarshalArticle(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"data":{
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

	defer func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d got: %d", http.StatusOK, resp.StatusCode)

		b, err := io.ReadAll(resp.Body)
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

func TestUnmarshalArticles(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"data":[
	{
		"type":"article",
		"id": "82180c8d-0ab6-4946-9298-61d3c8d13da4",
		"attributes": {
			"title": "This is the first article"
		}
	},
	{
		"type":"article",
		"id": "f3bbdbac-903c-4f2f-9721-600a1201ee41",
		"attributes": {
			"title": "This is the second article"
		}
	}
	]}`))
	req.Header.Set("Accept", JSONAPIContentType)
	req.Header.Set("Content-Type", JSONAPIContentType)

	type Article struct {
		ID    string `jsonapi:"primary,article" valid:"optional,uuid"`
		Title string `jsonapi:"attr,title" valid:"required"`
	}

	ok, articles := UnmarshalMany(rec, req, reflect.TypeOf(new(Article)))
	if !ok {
		t.Error("Un-marshalling many should have been ok")
	}

	resp := rec.Result()

	defer func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d got: %d", http.StatusOK, resp.StatusCode)

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		t.Error(string(b[:]))
	}

	if len(articles) != 2 {
		t.Errorf("Expected 2 articles, got %d", len(articles))
	}

	expected := []*Article{
		{
			ID:    "82180c8d-0ab6-4946-9298-61d3c8d13da4",
			Title: "This is the first article",
		},
		{
			ID:    "f3bbdbac-903c-4f2f-9721-600a1201ee41",
			Title: "This is the second article",
		},
	}

	for i := range articles {
		got, ok := articles[i].(*Article)
		if !ok {
			t.Errorf("Expected type *Article got: %T", articles[i])
		}

		if expected[i].ID != got.ID {
			t.Errorf("article.ID expected %q got: %q", expected[i].ID, got.ID)
		}

		if expected[i].Title != got.Title {
			t.Errorf("article.ID expected \"%s\" got: %q", expected[i].ID, got.Title)
		}
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

	defer func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d got: %d", http.StatusOK, resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expectation := `{"data":{"type":"articles","id":"cb855aff-f03c-4307-9a22-ab5fcc6b6d7c","attributes":{"title":"This is my first blog"}}}` + "\n"

	if expectation != string(b[:]) {
		t.Errorf("Expected %q got: %q", expectation, string(b[:]))
	}
}

func TestMarshalPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic in Marshal")
		}
	}()

	rec := httptest.NewRecorder()
	Marshal(rec, make(chan int), http.StatusOK)
}

type writer struct{}

func (writer) Header() http.Header {
	return http.Header{}
}

func (writer) WriteHeader(statusCode int) {
}

func (writer) Write(buf []byte) (int, error) {
	return 0, &net.OpError{
		Op:  "write",
		Err: fmt.Errorf("Test connection error"),
	}
}

func TestMarshalConnectionError(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatal("Was not expecting a panic")
		}
	}()

	rec := writer{}
	Marshal(rec, &struct{}{}, http.StatusOK)
}
