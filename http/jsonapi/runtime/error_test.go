// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestErrorMarshaling(t *testing.T) {
	testCases := []struct {
		name       string
		httpStatus int
		err        error
		result     []*Error
	}{
		{"Simple Error", http.StatusBadRequest, fmt.Errorf("Failed"), []*Error{
			&Error{Title: "Failed", Status: "400"},
		}},
		{"Other StatusCode", http.StatusUnauthorized, fmt.Errorf("Unauthorized"), []*Error{
			&Error{Title: "Unauthorized", Status: "401"},
		}},
		{"Error Object", http.StatusUnauthorized, Error{Title: "foo", Detail: "bar"}, []*Error{
			&Error{Title: "foo", Detail: "bar", Status: "401"},
		}},
		{"Error Object reference", http.StatusUnauthorized, &Error{Title: "foo", Detail: "bar"}, []*Error{
			&Error{Title: "foo", Detail: "bar", Status: "401"},
		}},
		{"Error Object List", http.StatusUnauthorized, Errors{
			&Error{Title: "foo", Detail: "bar"},
			&Error{Title: "foo2", Detail: "bar2"},
		}, []*Error{
			&Error{Title: "foo", Detail: "bar", Status: "401"},
			&Error{Title: "foo2", Detail: "bar2", Status: "401"},
		}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			WriteError(rec, testCase.httpStatus, testCase.err)

			resp := rec.Result()
			defer resp.Body.Close()

			if resp.StatusCode != testCase.httpStatus {
				t.Errorf("expected the response code %d got: %d", testCase.httpStatus, resp.StatusCode)
			}
			if ct := resp.Header.Get("Content-Type"); ct != JSONAPIContentType {
				t.Errorf("expected the response code %q got: %q", JSONAPIContentType, ct)
			}

			var errList errorObjects
			dec := json.NewDecoder(resp.Body)
			err := dec.Decode(&errList)
			if err != nil {
				t.Fatal(err)
			}

			if len(errList.List) != len(testCase.result) {
				t.Errorf("expected %d errors got %d", len(testCase.result), len(errList.List))
			}

			for i, errItem := range testCase.result {
				compareItem := errList.List[i]
				if !reflect.DeepEqual(errItem, compareItem) {
					t.Errorf("expected error #%d %#v to equal: %#v", i+1, errItem, compareItem)
				}
			}
		})
	}
}

func TestErrors(t *testing.T) {
	errs := Errors{
		&Error{Title: "foo", Detail: "bar"},
		&Error{Title: "foo2", Detail: "bar2"},
	}
	result := "foo\nfoo2"
	if errs.Error() != result {
		t.Errorf("expected %q got: %q", result, errs.Error())
	}
}

func TestError(t *testing.T) {
	err := Error{}
	err.setHTTPStatus(200)

	result := "200"
	if err.Status != result {
		t.Errorf("expected %q got: %q", result, err.Status)
	}
}
