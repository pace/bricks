// This file is originating from https://github.com/google/jsonapi/
// To this file the license conditions of LICENSE apply.

package jsonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestErrorObjectWritesExpectedErrorMessage(t *testing.T) {
	err := &ObjectError{Title: "Title test.", Detail: "Detail test."}

	var input error = err

	output := input.Error()

	if output != fmt.Sprintf("Error: %s %s\n", err.Title, err.Detail) {
		t.Fatal("Unexpected output.")
	}
}

func TestMarshalErrorsWritesTheExpectedPayload(t *testing.T) {
	marshalErrorsTableTasts := []struct {
		Title string
		In    []*ObjectError
		Out   map[string]any
	}{
		{
			Title: "TestFieldsAreSerializedAsNeeded",
			In:    []*ObjectError{{ID: "0", Title: "Test title.", Detail: "Test detail", Status: "http.StatusBadRequest", Code: "E1100"}},
			Out: map[string]any{"errors": []any{
				map[string]any{"id": "0", "title": "Test title.", "detail": "Test detail", "status": "http.StatusBadRequest", "code": "E1100"},
			}},
		},
		{
			Title: "TestMetaFieldIsSerializedProperly",
			In:    []*ObjectError{{Title: "Test title.", Detail: "Test detail", Meta: &map[string]any{"key": "val"}}},
			Out: map[string]any{"errors": []any{
				map[string]any{"title": "Test title.", "detail": "Test detail", "meta": map[string]any{"key": "val"}},
			}},
		},
	}
	for _, testRow := range marshalErrorsTableTasts {
		t.Run(testRow.Title, func(t *testing.T) {
			buffer, output := bytes.NewBuffer(nil), map[string]any{}

			var writer io.Writer = buffer

			_ = MarshalErrors(writer, testRow.In)

			if err := json.Unmarshal(buffer.Bytes(), &output); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(output, testRow.Out) {
				t.Fatalf("Expected: \n%#v \nto equal: \n%#v", output, testRow.Out)
			}
		})
	}
}
