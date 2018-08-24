// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package generator

import (
	"io/ioutil"
	"testing"

	"github.com/pmezard/go-difflib/difflib"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestGenerator(t *testing.T) {
	cases := []struct {
		title, path, source, pkg string
	}{
		{"PACE Fueling API", "./internal/fueling/open-api_test.go", "./internal/fueling/open-api.json", "fueling"},
		{"PACE Payment API", "./internal/pay/open-api_test.go", "./internal/pay/open-api.json", "pay"},
		{"PACE POI API", "./internal/poi/open-api_test.go", "./internal/poi/open-api.json", "poi"},
	}

	for _, testCase := range cases {
		t.Run(testCase.title, func(t *testing.T) {
			// read spec
			data, err := ioutil.ReadFile(testCase.source)
			if err != nil {
				t.Fatal(err)
			}

			// parse spec
			loader := openapi3.NewSwaggerLoader()
			schema, err := loader.LoadSwaggerFromData(data)
			if err != nil {
				t.Fatal(err)
			}

			// simple validation on the spec loading
			if schema.Info.Title != testCase.title {
				t.Errorf("Expected schema title to be %q got %q", testCase.title, schema.Info.Title)
			}
			// generate the go code in temp dir
			g := Generator{}

			result, err := g.BuildSchema(schema, testCase.path, testCase.pkg)
			if err != nil {
				t.Fatal(err)
			}

			expected, err := ioutil.ReadFile(testCase.path)
			if err != nil {
				t.Fatal(err)
			}

			if string(expected[:]) != result {
				diff := difflib.UnifiedDiff{
					A:        difflib.SplitLines(string(expected[:])),
					B:        difflib.SplitLines(result),
					FromFile: "Expected",
					ToFile:   "Generated",
					Context:  3,
				}
				text, _ := difflib.GetUnifiedDiffString(diff)
				t.Errorf(text)
			}
		})
	}
}
