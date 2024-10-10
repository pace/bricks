// This file is originating from https://github.com/google/jsonapi/
// To this file the license conditions of LICENSE apply.

package jsonapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshall_attrStringSlice(t *testing.T) {
	out := &Book{}
	tags := []string{"fiction", "sale"}
	data := map[string]any{
		"data": map[string]any{
			"type": "books",
			"id":   "1",
			"attributes": map[string]json.RawMessage{
				"tags": json.RawMessage(`["fiction", "sale"]`),
				"dec1": json.RawMessage(`"9.9999999999999999999"`),
				"dec2": json.RawMessage("9.9999999999999999999"),
				"dec3": json.RawMessage("10"),
			},
		},
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	if err := UnmarshalPayload(bytes.NewReader(b), out); err != nil {
		t.Fatal(err)
	}

	if e, a := len(tags), len(out.Tags); e != a {
		t.Fatalf("Was expecting %d tags, got %d", e, a)
	}

	sort.Stable(sort.StringSlice(tags))
	sort.Stable(sort.StringSlice(out.Tags))

	if out.Decimal1.String() != "9.9999999999999999999" {
		t.Fatalf("Expected json dec1 data to be %#v got: %#v", "9.9999999999999999999", out.Decimal1.String())
	}

	if out.Decimal2.String() != "9.9999999999999999999" {
		t.Fatalf("Expected json dec2 data to be %#v got: %#v", "9.9999999999999999999", out.Decimal2.String())
	}

	if out.Decimal3.String() != "10" {
		t.Fatalf("Expected json dec2 data to be %#v got: %#v", 10, out.Decimal3.String())
	}

	for i, tag := range tags {
		if e, a := tag, out.Tags[i]; e != a {
			t.Fatalf("At index %d, was expecting %s got %s", i, e, a)
		}
	}
}

func TestUnmarshall_MapStringSlice(t *testing.T) {
	tcs := []struct {
		name  string
		fail  bool
		input any
	}{
		{
			name: "succeed",
			fail: false,
			input: map[string]any{
				"data": map[string]any{
					"type": "books",
					"id":   "1",
					"attributes": map[string]json.RawMessage{
						"mss": json.RawMessage(`{"field": ["string1","string2"]}`),
					},
				},
			},
		},
		{
			name: "fail because slice contains numbers",
			fail: true,
			input: map[string]any{
				"data": map[string]any{
					"type": "books",
					"id":   "1",
					"attributes": map[string]json.RawMessage{
						"mss": json.RawMessage(`{"field": ["string1",1234,9.9]}`),
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out := &Book{}

			b, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatal(err)
			}

			err = UnmarshalPayload(bytes.NewReader(b), out)
			assert.Equal(t, tc.fail, err != nil)
		})
	}
}

func TestUnmarshalToStructWithPointerAttr(t *testing.T) {
	out := new(WithPointer)
	in := map[string]json.RawMessage{
		"name":      json.RawMessage(`"The name"`),
		"is-active": json.RawMessage(`true`),
		"int-val":   json.RawMessage(`8`),
		"float-val": json.RawMessage(`1.1`),
	}

	payload, err := sampleWithPointerPayload(in)
	require.NoError(t, err)

	if err := UnmarshalPayload(payload, out); err != nil {
		t.Fatal(err)
	}

	if *out.Name != "The name" {
		t.Fatalf("Error unmarshalling to string ptr")
	}

	if !*out.IsActive {
		t.Fatalf("Error unmarshalling to bool ptr")
	}

	if *out.IntVal != 8 {
		t.Fatalf("Error unmarshalling to int ptr")
	}

	if *out.FloatVal != 1.1 {
		t.Fatalf("Error unmarshalling to float ptr")
	}
}

func TestUnmarshalPayload_ptrsAllNil(t *testing.T) {
	out := new(WithPointer)
	if err := UnmarshalPayload(
		strings.NewReader(`{"data": {}}`), out); err != nil {
		t.Fatalf("Error unmarshalling to Foo")
	}

	if out.ID != nil {
		t.Fatalf("Error unmarshalling; expected ID ptr to be nil")
	}
}

func TestUnmarshalPayloadWithPointerID(t *testing.T) {
	out := new(WithPointer)
	attrs := map[string]json.RawMessage{}

	payload, err := sampleWithPointerPayload(attrs)
	require.NoError(t, err)

	if err := UnmarshalPayload(payload, out); err != nil {
		t.Fatalf("Error unmarshalling to Foo")
	}

	// these were present in the payload -- expect val to be not nil
	if out.ID == nil {
		t.Fatalf("Error unmarshalling; expected ID ptr to be not nil")
	}

	if e, a := uint64(2), *out.ID; e != a {
		t.Fatalf("Was expecting the ID to have a value of %d, got %d", e, a)
	}
}

func TestUnmarshalPayloadWithPointerAttr_AbsentVal(t *testing.T) {
	out := new(WithPointer)
	in := map[string]json.RawMessage{
		"name":      json.RawMessage(`"The name"`),
		"is-active": json.RawMessage(`true`),
	}

	payload, err := sampleWithPointerPayload(in)
	require.NoError(t, err)

	if err := UnmarshalPayload(payload, out); err != nil {
		t.Fatalf("Error unmarshalling to Foo")
	}

	// these were present in the payload -- expect val to be not nil
	if out.Name == nil || out.IsActive == nil {
		t.Fatalf("Error unmarshalling; expected ptr to be not nil")
	}

	// these were absent in the payload -- expect val to be nil
	if out.IntVal != nil || out.FloatVal != nil {
		t.Fatalf("Error unmarshalling; expected ptr to be nil")
	}
}

func TestUnmarshalToStructWithPointerAttr_BadType_bool(t *testing.T) {
	out := new(WithPointer)
	in := map[string]json.RawMessage{
		"name": json.RawMessage(`true`), // This is the wrong type.
	}
	expectedErrorMessage := "jsonapi: Can't unmarshal true (bool) to struct field `Name`, which is a pointer to `string`"

	payload, err := sampleWithPointerPayload(in)
	require.NoError(t, err)

	err = UnmarshalPayload(payload, out)
	if err == nil {
		t.Fatalf("Expected error due to invalid type.")
	}

	if err.Error() != expectedErrorMessage {
		t.Fatalf("Unexpected error message: %s", err.Error())
	}

	if _, ok := err.(UnsupportedPtrTypeError); !ok { //nolint:errorlint
		t.Fatalf("Unexpected error type: %s", reflect.TypeOf(err))
	}
}

func TestUnmarshalToStructWithPointerAttr_BadType_MapPtr(t *testing.T) {
	out := new(WithPointer)
	in := map[string]json.RawMessage{
		"name": json.RawMessage(`{"a": 5}`), // This is the wrong type.
	}
	expectedErrorMessage := "jsonapi: Can't unmarshal map[a:5] (map) to struct field `Name`, which is a pointer to `string`"

	payload, err := sampleWithPointerPayload(in)
	require.NoError(t, err)

	err = UnmarshalPayload(payload, out)
	if err == nil {
		t.Fatalf("Expected error due to invalid type.")
	}

	if err.Error() != expectedErrorMessage {
		t.Fatalf("Unexpected error message: %s", err.Error())
	}

	if _, ok := err.(UnsupportedPtrTypeError); !ok { //nolint:errorlint
		t.Fatalf("Unexpected error type: %s", reflect.TypeOf(err))
	}
}

func TestUnmarshalToStructWithPointerAttr_BadType_Struct(t *testing.T) {
	out := new(WithPointer)
	in := map[string]json.RawMessage{
		"name": json.RawMessage(`{"A": 5}`), // This is the wrong type.
	}
	expectedErrorMessage := "jsonapi: Can't unmarshal map[A:5] (map) to struct field `Name`, which is a pointer to `string`"

	payload, err := sampleWithPointerPayload(in)
	require.NoError(t, err)

	err = UnmarshalPayload(payload, out)
	if err == nil {
		t.Fatalf("Expected error due to invalid type.")
	}

	if err.Error() != expectedErrorMessage {
		t.Fatalf("Unexpected error message: %s", err.Error())
	}

	if _, ok := err.(UnsupportedPtrTypeError); !ok { //nolint:errorlint
		t.Fatalf("Unexpected error type: %s", reflect.TypeOf(err))
	}
}

func TestUnmarshalToStructWithPointerAttr_BadType_IntSlice(t *testing.T) {
	out := new(WithPointer)
	in := map[string]json.RawMessage{
		"name": json.RawMessage(`[4, 5]`), // This is the wrong type.
	}
	expectedErrorMessage := "jsonapi: Can't unmarshal [4 5] (slice) to struct field `Name`, which is a pointer to `string`"

	payload, err := sampleWithPointerPayload(in)
	require.NoError(t, err)

	err = UnmarshalPayload(payload, out)
	if err == nil {
		t.Fatalf("Expected error due to invalid type.")
	}

	if err.Error() != expectedErrorMessage {
		t.Fatalf("Unexpected error message: %s", err.Error())
	}

	if _, ok := err.(UnsupportedPtrTypeError); !ok { //nolint:errorlint
		t.Fatalf("Unexpected error type: %s", reflect.TypeOf(err))
	}
}

func TestStringPointerField(t *testing.T) {
	// Build Book payload
	description := "Hello World!"
	data := map[string]any{
		"data": map[string]any{
			"type": "books",
			"id":   "5",
			"attributes": map[string]any{
				"author":      "aren55555",
				"description": description,
				"isbn":        "",
			},
		},
	}

	payload, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	// Parse JSON API payload
	book := new(Book)
	if err := UnmarshalPayload(bytes.NewReader(payload), book); err != nil {
		t.Fatal(err)
	}

	if book.Description == nil {
		t.Fatal("Was not expecting a nil pointer for book.Description")
	}

	if expected, actual := description, *book.Description; expected != actual {
		t.Fatalf("Was expecting descript to be `%s`, got `%s`", expected, actual)
	}
}

func TestMalformedTag(t *testing.T) {
	out := new(BadModel)

	payload, err := samplePayload()
	require.NoError(t, err)

	err = UnmarshalPayload(payload, out)
	require.ErrorIs(t, err, ErrBadJSONAPIStructTag)
}

func TestUnmarshalInvalidJSON(t *testing.T) {
	in := strings.NewReader("{}")
	out := new(Blog)

	err := UnmarshalPayload(in, out)
	if err == nil {
		t.Fatalf("Did not error out the invalid JSON.")
	}
}

func TestUnmarshalInvalidJSON_BadType(t *testing.T) {
	badTypeTests := []struct {
		Field    string
		BadValue json.RawMessage
		Error    error
	}{ // The `Field` values here correspond to the `ModelBadTypes` jsonapi fields.
		{Field: "string_field", BadValue: json.RawMessage(`0`), Error: ErrUnknownFieldNumberType},                                                                   // Expected string.
		{Field: "float_field", BadValue: json.RawMessage(`"A string."`), Error: errors.New("got value \"A string.\" expected type float64: invalid type provided")}, // Expected float64.
		{Field: "time_field", BadValue: json.RawMessage(`"A string."`), Error: ErrInvalidTime},                                                                      // Expected int64.
		{Field: "time_ptr_field", BadValue: json.RawMessage(`"A string."`), Error: ErrInvalidTime},                                                                  // Expected *time / int64.
	}
	for i, test := range badTypeTests {
		t.Run(fmt.Sprintf("Test_%s", test.Field), func(t *testing.T) {
			out := new(ModelBadTypes)
			in := map[string]json.RawMessage{}
			in[test.Field] = test.BadValue
			expectedErrorMessage := test.Error.Error()

			payload, err := samplePayloadWithBadTypes(in)
			require.NoError(t, err)

			err = UnmarshalPayload(payload, out)
			if err == nil {
				t.Fatalf("(Test %d) Expected error due to invalid type.", i+1)
			}

			require.Equal(t, expectedErrorMessage, err.Error())
		})
	}
}

func TestUnmarshalSetsID(t *testing.T) {
	in, err := samplePayloadWithID()
	require.NoError(t, err)

	out := new(Blog)

	if err := UnmarshalPayload(in, out); err != nil {
		t.Fatal(err)
	}

	if out.ID != 2 {
		t.Fatalf("Did not set ID on dst interface")
	}
}

func TestUnmarshal_nonNumericID(t *testing.T) {
	data := samplePayloadWithoutIncluded()

	dataMap, ok := data["data"].(map[string]any)
	if !ok {
		t.Fatal("data is not a map")
	}

	dataMap["id"] = "non-numeric-id"

	payload, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	in := bytes.NewReader(payload)
	out := new(Post)

	err = UnmarshalPayload(in, out)
	require.ErrorIs(t, err, ErrBadJSONAPIID)
}

func TestUnmarshalSetsAttrs(t *testing.T) {
	out, err := unmarshalSamplePayload()
	if err != nil {
		t.Fatal(err)
	}

	if out.CreatedAt.IsZero() {
		t.Fatalf("Did not parse time")
	}

	if out.ViewCount != 1000 {
		t.Fatalf("View count not properly serialized")
	}
}

func TestUnmarshalParsesISO8601(t *testing.T) {
	payload := &OnePayload{
		Data: &Node{
			Type: "timestamps",
			Attributes: map[string]json.RawMessage{
				"timestamp": json.RawMessage(`"2016-08-17T08:27:12Z"`),
			},
		},
	}

	in := bytes.NewBuffer(nil)

	if err := json.NewEncoder(in).Encode(payload); err != nil {
		log.Fatal(err)
	}

	out := new(Timestamp)

	if err := UnmarshalPayload(in, out); err != nil {
		t.Fatal(err)
	}

	expected := time.Date(2016, 8, 17, 8, 27, 12, 0, time.UTC)

	if !out.Time.Equal(expected) {
		t.Fatal("Parsing the ISO8601 timestamp failed")
	}
}

func TestUnmarshalParsesISO8601TimePointer(t *testing.T) {
	payload := &OnePayload{
		Data: &Node{
			Type: "timestamps",
			Attributes: map[string]json.RawMessage{
				"next": json.RawMessage(`"2016-08-17T08:27:12Z"`),
			},
		},
	}

	in := bytes.NewBuffer(nil)

	if err := json.NewEncoder(in).Encode(payload); err != nil {
		t.Fatal(err)
	}

	out := new(Timestamp)

	if err := UnmarshalPayload(in, out); err != nil {
		t.Fatal(err)
	}

	expected := time.Date(2016, 8, 17, 8, 27, 12, 0, time.UTC)

	if !out.Next.Equal(expected) {
		t.Fatal("Parsing the ISO8601 timestamp failed")
	}
}

func TestUnmarshalInvalidISO8601(t *testing.T) {
	payload := &OnePayload{
		Data: &Node{
			Type: "timestamps",
			Attributes: map[string]json.RawMessage{
				"timestamp": json.RawMessage(`"17 Aug 16 08:027 MST"`),
			},
		},
	}

	in := bytes.NewBuffer(nil)

	if err := json.NewEncoder(in).Encode(payload); err != nil {
		t.Fatal(err)
	}

	out := new(Timestamp)

	err := UnmarshalPayload(in, out)
	require.ErrorIs(t, err, ErrInvalidISO8601)
}

func TestUnmarshalRelationshipsWithoutIncluded(t *testing.T) {
	data, err := json.Marshal(samplePayloadWithoutIncluded())
	if err != nil {
		t.Fatal(err)
	}

	in := bytes.NewReader(data)
	out := new(Post)

	if err := UnmarshalPayload(in, out); err != nil {
		t.Fatal(err)
	}

	// Verify each comment has at least an ID
	for _, comment := range out.Comments {
		if comment.ID == 0 {
			t.Fatalf("The comment did not have an ID")
		}
	}
}

func TestUnmarshalRelationships(t *testing.T) {
	out, err := unmarshalSamplePayload()
	if err != nil {
		t.Fatal(err)
	}

	if out.CurrentPost == nil {
		t.Fatalf("Current post was not materialized")
	}

	if out.CurrentPost.Title != "Bas" || out.CurrentPost.Body != "Fuubar" {
		t.Fatalf("Attributes were not set")
	}

	if len(out.Posts) != 2 {
		t.Fatalf("There should have been 2 posts")
	}
}

func TestUnmarshalNullRelationship(t *testing.T) {
	sample := map[string]any{
		"data": map[string]any{
			"type": "posts",
			"id":   "1",
			"attributes": map[string]any{
				"body":  "Hello",
				"title": "World",
			},
			"relationships": map[string]any{
				"latest_comment": map[string]any{
					"data": nil, // empty to-one relationship
				},
			},
		},
	}

	data, err := json.Marshal(sample)
	if err != nil {
		t.Fatal(err)
	}

	in := bytes.NewReader(data)
	out := new(Post)

	if err := UnmarshalPayload(in, out); err != nil {
		t.Fatal(err)
	}

	if out.LatestComment != nil {
		t.Fatalf("Latest Comment was not set to nil")
	}
}

func TestUnmarshalNullRelationshipInSlice(t *testing.T) {
	sample := map[string]any{
		"data": map[string]any{
			"type": "posts",
			"id":   "1",
			"attributes": map[string]any{
				"body":  "Hello",
				"title": "World",
			},
			"relationships": map[string]any{
				"comments": map[string]any{
					"data": []any{}, // empty to-many relationships
				},
			},
		},
	}

	data, err := json.Marshal(sample)
	if err != nil {
		t.Fatal(err)
	}

	in := bytes.NewReader(data)
	out := new(Post)

	if err := UnmarshalPayload(in, out); err != nil {
		t.Fatal(err)
	}

	if len(out.Comments) != 0 {
		t.Fatalf("Wrong number of comments; Comments should be empty")
	}
}

func TestUnmarshalNestedRelationships(t *testing.T) {
	out, err := unmarshalSamplePayload()
	if err != nil {
		t.Fatal(err)
	}

	if out.CurrentPost == nil {
		t.Fatalf("Current post was not materialized")
	}

	if out.CurrentPost.Comments == nil {
		t.Fatalf("Did not materialize nested records, comments")
	}

	if len(out.CurrentPost.Comments) != 2 {
		t.Fatalf("Wrong number of comments")
	}
}

func TestUnmarshalRelationshipsSerializedEmbedded(t *testing.T) {
	out := sampleSerializedEmbeddedTestModel()

	if out.CurrentPost == nil {
		t.Fatalf("Current post was not materialized")
	}

	if out.CurrentPost.Title != "Foo" || out.CurrentPost.Body != "Bar" {
		t.Fatalf("Attributes were not set")
	}

	if len(out.Posts) != 2 {
		t.Fatalf("There should have been 2 posts")
	}

	if out.Posts[0].LatestComment.Body != "foo" {
		t.Fatalf("The comment body was not set")
	}
}

func TestUnmarshalNestedRelationshipsEmbedded(t *testing.T) {
	out := bytes.NewBuffer(nil)
	if err := MarshalOnePayloadEmbedded(out, testModel()); err != nil {
		t.Fatal(err)
	}

	model := new(Blog)

	if err := UnmarshalPayload(out, model); err != nil {
		t.Fatal(err)
	}

	if model.CurrentPost == nil {
		t.Fatalf("Current post was not materialized")
	}

	if model.CurrentPost.Comments == nil {
		t.Fatalf("Did not materialize nested records, comments")
	}

	if len(model.CurrentPost.Comments) != 2 {
		t.Fatalf("Wrong number of comments")
	}

	if model.CurrentPost.Comments[0].Body != "foo" {
		t.Fatalf("Comment body not set")
	}
}

func TestUnmarshalRelationshipsSideloaded(t *testing.T) {
	payload := samplePayloadWithSideloaded()
	out := new(Blog)

	if err := UnmarshalPayload(payload, out); err != nil {
		t.Fatal(err)
	}

	if out.CurrentPost == nil {
		t.Fatalf("Current post was not materialized")
	}

	if out.CurrentPost.Title != "Foo" || out.CurrentPost.Body != "Bar" {
		t.Fatalf("Attributes were not set")
	}

	if len(out.Posts) != 2 {
		t.Fatalf("There should have been 2 posts")
	}
}

func TestUnmarshalNestedRelationshipsSideloaded(t *testing.T) {
	payload := samplePayloadWithSideloaded()
	out := new(Blog)

	if err := UnmarshalPayload(payload, out); err != nil {
		t.Fatal(err)
	}

	if out.CurrentPost == nil {
		t.Fatalf("Current post was not materialized")
	}

	if out.CurrentPost.Comments == nil {
		t.Fatalf("Did not materialize nested records, comments")
	}

	if len(out.CurrentPost.Comments) != 2 {
		t.Fatalf("Wrong number of comments")
	}

	if out.CurrentPost.Comments[0].Body != "foo" {
		t.Fatalf("Comment body not set")
	}
}

func TestUnmarshalNestedRelationshipsEmbedded_withClientIDs(t *testing.T) {
	model := new(Blog)

	payload, err := samplePayload()
	require.NoError(t, err)

	if err := UnmarshalPayload(payload, model); err != nil {
		t.Fatal(err)
	}

	if model.Posts[0].ClientID == "" {
		t.Fatalf("ClientID not set from request on related record")
	}
}

func unmarshalSamplePayload() (*Blog, error) {
	in, err := samplePayload()
	if err != nil {
		return nil, err
	}

	out := new(Blog)

	if err := UnmarshalPayload(in, out); err != nil {
		return nil, err
	}

	return out, nil
}

func TestUnmarshalManyPayload(t *testing.T) {
	sample := map[string]any{
		"data": []any{
			map[string]any{
				"type": "posts",
				"id":   "1",
				"attributes": map[string]any{
					"body":  "First",
					"title": "Post",
				},
			},
			map[string]any{
				"type": "posts",
				"id":   "2",
				"attributes": map[string]any{
					"body":  "Second",
					"title": "Post",
				},
			},
		},
	}

	data, err := json.Marshal(sample)
	if err != nil {
		t.Fatal(err)
	}

	in := bytes.NewReader(data)

	posts, err := UnmarshalManyPayload(in, reflect.TypeOf(new(Post)))
	if err != nil {
		t.Fatal(err)
	}

	if len(posts) != 2 {
		t.Fatal("Wrong number of posts")
	}

	for _, p := range posts {
		_, ok := p.(*Post)
		if !ok {
			t.Fatal("Was expecting a Post")
		}
	}
}

func TestManyPayload_withLinks(t *testing.T) {
	firstPageURL := "http://somesite.com/movies?page[limit]=50&page[offset]=50"
	prevPageURL := "http://somesite.com/movies?page[limit]=50&page[offset]=0"
	nextPageURL := "http://somesite.com/movies?page[limit]=50&page[offset]=100"
	lastPageURL := "http://somesite.com/movies?page[limit]=50&page[offset]=500"

	sample := map[string]any{
		"data": []any{
			map[string]any{
				"type": "posts",
				"id":   "1",
				"attributes": map[string]any{
					"body":  "First",
					"title": "Post",
				},
			},
			map[string]any{
				"type": "posts",
				"id":   "2",
				"attributes": map[string]any{
					"body":  "Second",
					"title": "Post",
				},
			},
		},
		"links": map[string]any{
			KeyFirstPage:    firstPageURL,
			KeyPreviousPage: prevPageURL,
			KeyNextPage:     nextPageURL,
			KeyLastPage:     lastPageURL,
		},
	}

	data, err := json.Marshal(sample)
	if err != nil {
		t.Fatal(err)
	}

	in := bytes.NewReader(data)

	payload := new(ManyPayload)
	if err = json.NewDecoder(in).Decode(payload); err != nil {
		t.Fatal(err)
	}

	if payload.Links == nil {
		t.Fatal("Was expecting a non nil ptr Link field")
	}

	links := *payload.Links

	first, ok := links[KeyFirstPage]
	if !ok {
		t.Fatal("Was expecting a non nil ptr Link field")
	}

	if e, a := firstPageURL, first; e != a {
		t.Fatalf("Was expecting links.%s to have a value of %s, got %s", KeyFirstPage, e, a)
	}

	prev, ok := links[KeyPreviousPage]
	if !ok {
		t.Fatal("Was expecting a non nil ptr Link field")
	}

	if e, a := prevPageURL, prev; e != a {
		t.Fatalf("Was expecting links.%s to have a value of %s, got %s", KeyPreviousPage, e, a)
	}

	next, ok := links[KeyNextPage]
	if !ok {
		t.Fatal("Was expecting a non nil ptr Link field")
	}

	if e, a := nextPageURL, next; e != a {
		t.Fatalf("Was expecting links.%s to have a value of %s, got %s", KeyNextPage, e, a)
	}

	last, ok := links[KeyLastPage]
	if !ok {
		t.Fatal("Was expecting a non nil ptr Link field")
	}

	if e, a := lastPageURL, last; e != a {
		t.Fatalf("Was expecting links.%s to have a value of %s, got %s", KeyLastPage, e, a)
	}
}

func TestUnmarshalCustomTypeAttributes(t *testing.T) {
	customInt := CustomIntType(5)
	customFloat := CustomFloatType(1.5)
	customString := CustomStringType("Test")

	data := map[string]any{
		"data": map[string]any{
			"type": "customtypes",
			"id":   "1",
			"attributes": map[string]json.RawMessage{
				"int":        json.RawMessage("5"),
				"intptr":     json.RawMessage("5"),
				"intptrnull": json.RawMessage("null"),
				"float":      json.RawMessage("1.5"),
				"string":     json.RawMessage(`"Test"`),
			},
		},
	}

	payload, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	// Parse JSON API payload
	customAttributeTypes := new(CustomAttributeTypes)
	if err := UnmarshalPayload(bytes.NewReader(payload), customAttributeTypes); err != nil {
		t.Fatal(err)
	}

	if expected, actual := customInt, customAttributeTypes.Int; expected != actual {
		t.Fatalf("Was expecting custom int to be `%d`, got `%d`", expected, actual)
	}

	if expected, actual := customInt, *customAttributeTypes.IntPtr; expected != actual {
		t.Fatalf("Was expecting custom int pointer to be `%d`, got `%d`", expected, actual)
	}

	if customAttributeTypes.IntPtrNull != nil {
		t.Fatalf("Was expecting custom int pointer to be <nil>, got `%d`", customAttributeTypes.IntPtrNull)
	}

	if expected, actual := customFloat, customAttributeTypes.Float; expected != actual {
		t.Fatalf("Was expecting custom float to be `%f`, got `%f`", expected, actual)
	}

	if expected, actual := customString, customAttributeTypes.String; expected != actual {
		t.Fatalf("Was expecting custom string to be `%s`, got `%s`", expected, actual)
	}
}

func TestUnmarshalCustomTypeAttributes_ErrInvalidType(t *testing.T) {
	data := map[string]any{
		"data": map[string]any{
			"type": "customtypes",
			"id":   "1",
			"attributes": map[string]json.RawMessage{
				"int":        json.RawMessage(`"bad"`),
				"intptr":     json.RawMessage(`5`),
				"intptrnull": json.RawMessage(`null`),
				"float":      json.RawMessage(`1.5`),
				"string":     json.RawMessage(`"Test"`),
			},
		},
	}

	payload, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	// Parse JSON API payload
	customAttributeTypes := new(CustomAttributeTypes)

	err = UnmarshalPayload(bytes.NewReader(payload), customAttributeTypes)
	if err == nil {
		t.Fatal("Expected an error unmarshalling the payload due to type mismatch, got none")
	}

	e := errors.New("got value \"bad\" expected type jsonapi.CustomIntType: invalid type provided")
	if err.Error() != e.Error() {
		t.Fatalf("Expected error to be %q,\nwas %q", e, err)
	}
}

func samplePayloadWithoutIncluded() map[string]any {
	return map[string]any{
		"data": map[string]any{
			"type": "posts",
			"id":   "1",
			"attributes": map[string]json.RawMessage{
				"body":  json.RawMessage(`"Hello"`),
				"title": json.RawMessage(`"World"`),
			},
			"relationships": map[string]any{
				"comments": map[string]any{
					"data": []any{
						map[string]any{
							"type": "comments",
							"id":   "123",
						},
						map[string]any{
							"type": "comments",
							"id":   "456",
						},
					},
				},
				"latest_comment": map[string]any{
					"data": map[string]any{
						"type": "comments",
						"id":   "55555",
					},
				},
			},
		},
	}
}

func samplePayload() (io.Reader, error) {
	payload := &OnePayload{
		Data: &Node{
			Type: "blogs",
			Attributes: map[string]json.RawMessage{
				"title":      json.RawMessage(`"New blog"`),
				"created_at": json.RawMessage(`1436216820`),
				"view_count": json.RawMessage(`1000`),
			},
			Relationships: map[string]any{
				"posts": &RelationshipManyNode{
					Data: []*Node{
						{
							Type: "posts",
							Attributes: map[string]json.RawMessage{
								"title": json.RawMessage(`"Foo"`),
								"body":  json.RawMessage(`"Bar"`),
							},
							ClientID: "1",
						},
						{
							Type: "posts",
							Attributes: map[string]json.RawMessage{
								"title": json.RawMessage(`"X"`),
								"body":  json.RawMessage(`"Y"`),
							},
							ClientID: "2",
						},
					},
				},
				"current_post": &RelationshipOneNode{
					Data: &Node{
						Type: "posts",
						Attributes: map[string]json.RawMessage{
							"title": json.RawMessage(`"Bas"`),
							"body":  json.RawMessage(`"Fuubar"`),
						},
						ClientID: "3",
						Relationships: map[string]any{
							"comments": &RelationshipManyNode{
								Data: []*Node{
									{
										Type: "comments",
										Attributes: map[string]json.RawMessage{
											"body": json.RawMessage(`"Great post!"`),
										},
										ClientID: "4",
									},
									{
										Type: "comments",
										Attributes: map[string]json.RawMessage{
											"body": json.RawMessage(`"Needs some work!"`),
										},
										ClientID: "5",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	out := bytes.NewBuffer(nil)

	if err := json.NewEncoder(out).Encode(payload); err != nil {
		return nil, err
	}

	return out, nil
}

func samplePayloadWithID() (io.Reader, error) {
	payload := &OnePayload{
		Data: &Node{
			ID:   "2",
			Type: "blogs",
			Attributes: map[string]json.RawMessage{
				"title":      json.RawMessage(`"New blog"`),
				"view_count": json.RawMessage(`1000`),
			},
		},
	}

	out := bytes.NewBuffer(nil)

	if err := json.NewEncoder(out).Encode(payload); err != nil {
		return nil, err
	}

	return out, nil
}

func samplePayloadWithBadTypes(m map[string]json.RawMessage) (io.Reader, error) {
	payload := &OnePayload{
		Data: &Node{
			ID:         "2",
			Type:       "badtypes",
			Attributes: m,
		},
	}

	out := bytes.NewBuffer(nil)

	if err := json.NewEncoder(out).Encode(payload); err != nil {
		return nil, err
	}

	return out, nil
}

func sampleWithPointerPayload(m map[string]json.RawMessage) (io.Reader, error) {
	payload := &OnePayload{
		Data: &Node{
			ID:         "2",
			Type:       "with-pointers",
			Attributes: m,
		},
	}

	out := bytes.NewBuffer(nil)

	if err := json.NewEncoder(out).Encode(payload); err != nil {
		return nil, err
	}

	return out, nil
}

func testModel() *Blog {
	return &Blog{
		ID:        5,
		ClientID:  "1",
		Title:     "Title 1",
		CreatedAt: time.Now(),
		Posts: []*Post{
			{
				ID:    1,
				Title: "Foo",
				Body:  "Bar",
				Comments: []*Comment{
					{
						ID:   1,
						Body: "foo",
					},
					{
						ID:   2,
						Body: "bar",
					},
				},
				LatestComment: &Comment{
					ID:   1,
					Body: "foo",
				},
			},
			{
				ID:    2,
				Title: "Fuubar",
				Body:  "Bas",
				Comments: []*Comment{
					{
						ID:   1,
						Body: "foo",
					},
					{
						ID:   3,
						Body: "bas",
					},
				},
				LatestComment: &Comment{
					ID:   1,
					Body: "foo",
				},
			},
		},
		CurrentPost: &Post{
			ID:    1,
			Title: "Foo",
			Body:  "Bar",
			Comments: []*Comment{
				{
					ID:   1,
					Body: "foo",
				},
				{
					ID:   2,
					Body: "bar",
				},
			},
			LatestComment: &Comment{
				ID:   1,
				Body: "foo",
			},
		},
	}
}

func samplePayloadWithSideloaded() io.Reader {
	testModel := testModel()

	out := bytes.NewBuffer(nil)

	if err := MarshalPayload(out, testModel); err != nil {
		panic(err)
	}

	return out
}

func sampleSerializedEmbeddedTestModel() *Blog {
	out := bytes.NewBuffer(nil)

	if err := MarshalOnePayloadEmbedded(out, testModel()); err != nil {
		panic(err)
	}

	blog := new(Blog)

	if err := UnmarshalPayload(out, blog); err != nil {
		panic(err)
	}

	return blog
}

func TestUnmarshalNestedStructPtr(t *testing.T) {
	type Director struct {
		Firstname string `jsonapi:"attr,firstname"`
		Surname   string `jsonapi:"attr,surname"`
	}

	type Movie struct {
		ID       string    `jsonapi:"primary,movies"`
		Name     string    `jsonapi:"attr,name"`
		Director *Director `jsonapi:"attr,director"`
	}

	sample := map[string]any{
		"data": map[string]any{
			"type": "movies",
			"id":   "123",
			"attributes": map[string]any{
				"name": "The Shawshank Redemption",
				"director": map[string]any{
					"firstname": "Frank",
					"surname":   "Darabont",
				},
			},
		},
	}

	data, err := json.Marshal(sample)
	if err != nil {
		t.Fatal(err)
	}

	in := bytes.NewReader(data)
	out := new(Movie)

	if err := UnmarshalPayload(in, out); err != nil {
		t.Fatal(err)
	}

	if out.Name != "The Shawshank Redemption" {
		t.Fatalf("expected out.Name to be `The Shawshank Redemption`, but got `%s`", out.Name)
	}

	if out.Director.Firstname != "Frank" {
		t.Fatalf("expected out.Director.Firstname to be `Frank`, but got `%s`", out.Director.Firstname)
	}

	if out.Director.Surname != "Darabont" {
		t.Fatalf("expected out.Director.Surname to be `Darabont`, but got `%s`", out.Director.Surname)
	}
}

func TestUnmarshalNestedStruct(t *testing.T) {
	boss := map[string]any{
		"firstname": "Hubert",
		"surname":   "Farnsworth",
		"age":       176,
		"hired-at":  "2016-08-17T08:27:12Z",
	}

	sample := map[string]any{
		"data": map[string]any{
			"type": "companies",
			"id":   "123",
			"attributes": map[string]any{
				"name":       "Planet Express",
				"boss":       boss,
				"founded-at": "2016-08-17T08:27:12Z",
				"teams": []map[string]any{
					{
						"name": "Dev",
						"members": []map[string]any{
							{"firstname": "Sean"},
							{"firstname": "Iz"},
						},
						"leader": map[string]any{"firstname": "Iz"},
					},
					{
						"name": "DxE",
						"members": []map[string]any{
							{"firstname": "Akshay"},
							{"firstname": "Peri"},
						},
						"leader": map[string]any{"firstname": "Peri"},
					},
				},
			},
		},
	}

	data, err := json.Marshal(sample)
	if err != nil {
		t.Fatal(err)
	}

	in := bytes.NewReader(data)
	out := new(Company)

	if err := UnmarshalPayload(in, out); err != nil {
		t.Fatal(err)
	}

	if out.Boss.Firstname != "Hubert" {
		t.Fatalf("expected `Hubert` at out.Boss.Firstname, but got `%s`", out.Boss.Firstname)
	}

	if out.Boss.Age != 176 {
		t.Fatalf("expected `176` at out.Boss.Age, but got `%d`", out.Boss.Age)
	}

	if out.Boss.HiredAt.IsZero() {
		t.Fatalf("expected out.Boss.HiredAt to be zero, but got `%t`", out.Boss.HiredAt.IsZero())
	}

	if len(out.Teams) != 2 {
		t.Fatalf("expected len(out.Teams) to be 2, but got `%d`", len(out.Teams))
	}

	if out.Teams[0].Name != "Dev" {
		t.Fatalf("expected out.Teams[0].Name to be `Dev`, but got `%s`", out.Teams[0].Name)
	}

	if out.Teams[1].Name != "DxE" {
		t.Fatalf("expected out.Teams[1].Name to be `DxE`, but got `%s`", out.Teams[1].Name)
	}

	if len(out.Teams[0].Members) != 2 {
		t.Fatalf("expected len(out.Teams[0].Members) to be 2, but got `%d`", len(out.Teams[0].Members))
	}

	if len(out.Teams[1].Members) != 2 {
		t.Fatalf("expected len(out.Teams[1].Members) to be 2, but got `%d`", len(out.Teams[1].Members))
	}

	if out.Teams[0].Members[0].Firstname != "Sean" {
		t.Fatalf("expected out.Teams[0].Members[0].Firstname to be `Sean`, but got `%s`", out.Teams[0].Members[0].Firstname)
	}

	if out.Teams[0].Members[1].Firstname != "Iz" {
		t.Fatalf("expected out.Teams[0].Members[1].Firstname to be `Iz`, but got `%s`", out.Teams[0].Members[1].Firstname)
	}

	if out.Teams[1].Members[0].Firstname != "Akshay" {
		t.Fatalf("expected out.Teams[1].Members[0].Firstname to be `Akshay`, but got `%s`", out.Teams[1].Members[0].Firstname)
	}

	if out.Teams[1].Members[1].Firstname != "Peri" {
		t.Fatalf("expected out.Teams[1].Members[1].Firstname to be `Peri`, but got `%s`", out.Teams[1].Members[1].Firstname)
	}

	if out.Teams[0].Leader.Firstname != "Iz" {
		t.Fatalf("expected out.Teams[0].Leader.Firstname to be `Iz`, but got `%s`", out.Teams[0].Leader.Firstname)
	}

	if out.Teams[1].Leader.Firstname != "Peri" {
		t.Fatalf("expected out.Teams[1].Leader.Firstname to be `Peri`, but got `%s`", out.Teams[1].Leader.Firstname)
	}
}

func TestUnmarshalNestedStructSlice(t *testing.T) {
	fry := map[string]any{
		"firstname": "Philip J.",
		"surname":   "Fry",
		"age":       25,
		"hired-at":  "2016-08-17T08:27:12Z",
	}

	bender := map[string]any{
		"firstname": "Bender Bending",
		"surname":   "Rodriguez",
		"age":       19,
		"hired-at":  "2016-08-17T08:27:12Z",
	}

	deliveryCrew := map[string]any{
		"name": "Delivery Crew",
		"members": []any{
			fry,
			bender,
		},
	}

	sample := map[string]any{
		"data": map[string]any{
			"type": "companies",
			"id":   "123",
			"attributes": map[string]any{
				"name": "Planet Express",
				"teams": []any{
					deliveryCrew,
				},
			},
		},
	}

	data, err := json.Marshal(sample)
	if err != nil {
		t.Fatal(err)
	}

	in := bytes.NewReader(data)
	out := new(Company)

	if err := UnmarshalPayload(in, out); err != nil {
		t.Fatal(err)
	}

	if out.Teams[0].Name != "Delivery Crew" {
		t.Fatalf("Nested struct not unmarshalled: Expected `Delivery Crew` but got `%s`", out.Teams[0].Name)
	}

	if len(out.Teams[0].Members) != 2 {
		t.Fatalf("Nested struct not unmarshalled: Expected to have `2` Members but got `%d`",
			len(out.Teams[0].Members))
	}

	if out.Teams[0].Members[0].Firstname != "Philip J." {
		t.Fatalf("Nested struct not unmarshalled: Expected `Philip J.` but got `%s`",
			out.Teams[0].Members[0].Firstname)
	}
}
