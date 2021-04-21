// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package generator

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

func TestGenerator(t *testing.T) {
	cases := []struct {
		title, path, source, pkg string
	}{
		{"PACE Fueling API", "./internal/fueling/open-api_test.go", "./internal/fueling/open-api.json", "fueling"},
		{"PACE Payment API", "./internal/pay/open-api_test.go", "./internal/pay/open-api.json", "pay"},
		{"PACE POI API", "./internal/poi/open-api_test.go", "./internal/poi/open-api.json", "poi"},
		{"Articles Test Service API", "./internal/articles/open-api_test.go", "./internal/articles/open-api.json", "articles"},
		{"Security Test API", "./internal/securitytest/open-api_test.go", "./internal/securitytest/open-api.json", "securitytest"},
	}

	for _, testCase := range cases {
		t.Run(testCase.title, func(t *testing.T) {
			expected, err := ioutil.ReadFile(testCase.path)
			if err != nil {
				t.Fatal(err)
			}

			g := Generator{}
			result, err := g.BuildSource(testCase.source, filepath.Dir(testCase.pkg), filepath.Base(testCase.pkg))
			if err != nil {
				t.Fatal(err)
			}
			if os.Getenv("PACE_TEST_GENERATOR_WRITE") != "" {
				f, err := os.Create(fmt.Sprintf("testout/test.%s.out.go", testCase.pkg))
				if err != nil {
					t.Fatal(err)
				}
				_, err = f.WriteString(result)
				if err != nil {
					t.Fatal(err)
				}
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
