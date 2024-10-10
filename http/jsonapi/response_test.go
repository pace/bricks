// This file is originating from https://github.com/google/jsonapi/
// To this file the license conditions of LICENSE apply.

package jsonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/pace/bricks/pkg/isotime"
)

func TestMarshalPayload(t *testing.T) {
	d, e := decimal.NewFromString("9.9999999999999999999")
	if e != nil {
		panic(e)
	}

	book := &Book{ID: 1, Decimal1: d}
	books := []*Book{book, {ID: 2}}

	var jsonData map[string]any

	// One
	out1 := bytes.NewBuffer(nil)

	if err := MarshalPayload(out1, book); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(out1.String(), `"9.9999999999999999999"`) {
		t.Fatalf("decimals should be encoded as number, got: %q", out1.String())
	}

	if err := json.Unmarshal(out1.Bytes(), &jsonData); err != nil {
		t.Fatal(err)
	}

	if _, ok := jsonData["data"].(map[string]any); !ok {
		t.Fatalf("data key did not contain an Hash/Dict/Map")
	}

	fmt.Println(out1.String())

	// Many
	out2 := bytes.NewBuffer(nil)

	if err := MarshalPayload(out2, books); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(out2.Bytes(), &jsonData); err != nil {
		t.Fatal(err)
	}

	if _, ok := jsonData["data"].([]any); !ok {
		t.Fatalf("data key did not contain an Array")
	}
}

func TestMarshalPayloadWithNulls(t *testing.T) {
	books := []*Book{nil, {ID: 101}, nil}

	var jsonData map[string]any

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, books); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(out.Bytes(), &jsonData); err != nil {
		t.Fatal(err)
	}

	raw, ok := jsonData["data"]
	if !ok {
		t.Fatalf("data key does not exist")
	}

	arr, ok := raw.([]any)
	if !ok {
		t.Fatalf("data is not an Array")
	}

	for i := range len(arr) {
		if books[i] == nil && arr[i] != nil ||
			books[i] != nil && arr[i] == nil {
			t.Fatalf("restored data is not equal to source")
		}
	}
}

func TestMarshal_attrStringSlice(t *testing.T) {
	tags := []string{"fiction", "sale"}
	b := &Book{ID: 1, Tags: tags}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, b); err != nil {
		t.Fatal(err)
	}

	var jsonData map[string]any
	if err := json.Unmarshal(out.Bytes(), &jsonData); err != nil {
		t.Fatal(err)
	}

	dataMap, ok := jsonData["data"].(map[string]any)
	if !ok {
		t.Fatal("data was not a map")
	}

	attributesMap, ok := dataMap["attributes"].(map[string]any)
	if !ok {
		t.Fatal("data.attributes was not a map")
	}

	jsonTags, ok := attributesMap["tags"].([]any)
	if !ok {
		t.Fatal("data.attributes.tags was not a slice")
	}

	if e, a := len(tags), len(jsonTags); e != a {
		t.Fatalf("Was expecting tags of length %d got %d", e, a)
	}

	// Convert from []any to []string
	jsonTagsStrings := []string{}

	for _, tag := range jsonTags {
		s, ok := tag.(string)
		if !ok {
			t.Fatalf("Was expecting tag to be a string, got %T", tag)
		}

		jsonTagsStrings = append(jsonTagsStrings, s)
	}

	// Sort both
	sort.Stable(sort.StringSlice(jsonTagsStrings))
	sort.Stable(sort.StringSlice(tags))

	for i, tag := range tags {
		if e, a := tag, jsonTagsStrings[i]; e != a {
			t.Fatalf("At index %d, was expecting %s got %s", i, e, a)
		}
	}
}

func TestWithoutOmitsEmptyAnnotationOnRelation(t *testing.T) {
	blog := &Blog{}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, blog); err != nil {
		t.Fatal(err)
	}

	var jsonData map[string]any
	if err := json.Unmarshal(out.Bytes(), &jsonData); err != nil {
		t.Fatal(err)
	}

	dataMap, ok := jsonData["data"].(map[string]any)
	if !ok {
		t.Fatal("data was not a map")
	}

	relationships, ok := dataMap["relationships"].(map[string]any)
	if !ok {
		t.Fatal("data.relationships was not a map")
	}

	// Verifiy the "posts" relation was an empty array
	posts, ok := relationships["posts"]
	if !ok {
		t.Fatal("Was expecting the data.relationships.posts key/value to have been present")
	}

	postsMap, ok := posts.(map[string]any)
	if !ok {
		t.Fatal("data.relationships.posts was not a map")
	}

	postsData, ok := postsMap["data"]
	if !ok {
		t.Fatal("Was expecting the data.relationships.posts.data key/value to have been present")
	}

	postsDataSlice, ok := postsData.([]any)
	if !ok {
		t.Fatal("data.relationships.posts.data was not a slice []")
	}

	if len(postsDataSlice) != 0 {
		t.Fatal("Was expecting the data.relationships.posts.data value to have been an empty array []")
	}

	// Verifiy the "current_post" was a null
	currentPost, postExists := relationships["current_post"]
	if !postExists {
		t.Fatal("Was expecting the data.relationships.current_post key/value to have NOT been omitted")
	}

	currentPostMap, ok := currentPost.(map[string]any)
	if !ok {
		t.Fatal("data.relationships.current_post was not a map")
	}

	currentPostData, ok := currentPostMap["data"]
	if !ok {
		t.Fatal("Was expecting the data.relationships.current_post.data key/value to have been present")
	}

	if currentPostData != nil {
		t.Fatal("Was expecting the data.relationships.current_post.data value to have been nil/null")
	}
}

func TestWithOmitsEmptyAnnotationOnRelation(t *testing.T) {
	type BlogOptionalPosts struct {
		ID          int     `jsonapi:"primary,blogs"`
		Title       string  `jsonapi:"attr,title"`
		Posts       []*Post `jsonapi:"relation,posts,omitempty"`
		CurrentPost *Post   `jsonapi:"relation,current_post,omitempty"`
	}

	blog := &BlogOptionalPosts{ID: 999}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, blog); err != nil {
		t.Fatal(err)
	}

	var jsonData map[string]any
	if err := json.Unmarshal(out.Bytes(), &jsonData); err != nil {
		t.Fatal(err)
	}

	payload, ok := jsonData["data"].(map[string]any)
	if !ok {
		t.Fatal("data was not a map")
	}

	// Verify relationship was NOT set
	if val, exists := payload["relationships"]; exists {
		t.Fatalf("Was expecting the data.relationships key/value to have been empty - it was not and had a value of %v", val)
	}
}

func TestWithOmitsEmptyAnnotationOnRelation_MixedData(t *testing.T) {
	type BlogOptionalPosts struct {
		ID          int     `jsonapi:"primary,blogs"`
		Title       string  `jsonapi:"attr,title"`
		Posts       []*Post `jsonapi:"relation,posts,omitempty"`
		CurrentPost *Post   `jsonapi:"relation,current_post,omitempty"`
	}

	blog := &BlogOptionalPosts{
		ID: 999,
		CurrentPost: &Post{
			ID: 123,
		},
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, blog); err != nil {
		t.Fatal(err)
	}

	var jsonData map[string]any
	if err := json.Unmarshal(out.Bytes(), &jsonData); err != nil {
		t.Fatal(err)
	}

	payload, ok := jsonData["data"].(map[string]any)
	if !ok {
		t.Fatal("data was not a map")
	}

	// Verify relationship was set
	if _, exists := payload["relationships"]; !exists {
		t.Fatal("Was expecting the data.relationships key/value to have NOT been empty")
	}

	relationships, ok := payload["relationships"].(map[string]any)
	if !ok {
		t.Fatal("data.relationships was not a map")
	}

	// Verify the relationship was not omitted, and is not null
	if val, exists := relationships["current_post"]; !exists {
		t.Fatal("Was expecting the data.relationships.current_post key/value to have NOT been omitted")
	} else {
		valMap, ok := val.(map[string]any)
		if !ok {
			t.Fatal("Was expecting the data.relationships.current_post value to have been a map")
		}

		if valMap["data"] == nil {
			t.Fatal("Was expecting the data.relationships.current_post value to have NOT been nil/null")
		}
	}
}

func TestWithOmitsEmptyAnnotationOnAttribute(t *testing.T) {
	type Phone struct {
		Number string `json:"number"`
	}

	type Address struct {
		City   string `json:"city"`
		Street string `json:"street"`
	}

	type Tags map[string]int

	type Author struct {
		ID      int              `jsonapi:"primary,authors"`
		Name    string           `jsonapi:"attr,title"`
		Phones  []*Phone         `jsonapi:"attr,phones,omitempty"`
		Address *Address         `jsonapi:"attr,address,omitempty"`
		Tags    Tags             `jsonapi:"attr,tags,omitempty"`
		Account *decimal.Decimal `jsonapi:"attr,account,omitempty"`
	}

	author := &Author{
		ID:      999,
		Name:    "Igor",
		Phones:  nil,                        // should be omitted
		Address: nil,                        // should be omitted
		Tags:    Tags{"dogs": 1, "cats": 2}, // should not be omitted
		Account: &decimal.Zero,
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, author); err != nil {
		t.Fatal(err)
	}

	var jsonData map[string]any
	if err := json.Unmarshal(out.Bytes(), &jsonData); err != nil {
		t.Fatal(err)
	}

	// Verify that there is no field "phones" in attributes
	payload, ok := jsonData["data"].(map[string]any)
	if !ok {
		t.Fatal("data was not a map")
	}

	attributes, ok := payload["attributes"].(map[string]any)
	if !ok {
		t.Fatal("Was expecting the data.attributes key/value to have been a map")
	}

	if _, ok := attributes["title"]; !ok {
		t.Fatal("Was expecting the data.attributes.title to have NOT been omitted")
	}

	if _, ok := attributes["phones"]; ok {
		t.Fatal("Was expecting the data.attributes.phones to have been omitted")
	}

	if _, ok := attributes["address"]; ok {
		t.Fatal("Was expecting the data.attributes.phones to have been omitted")
	}

	if _, ok := attributes["tags"]; !ok {
		t.Fatal("Was expecting the data.attributes.tags to have NOT been omitted")
	}

	if _, ok := attributes["account"]; !ok {
		t.Fatal("Was expecting the data.attributes.account to have NOT been omitted")
	}
}

func TestMarshalIDPtr(t *testing.T) {
	id, make, model := "123e4567-e89b-12d3-a456-426655440000", "Ford", "Mustang"
	car := &Car{
		ID:    &id,
		Make:  &make,
		Model: &model,
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, car); err != nil {
		t.Fatal(err)
	}

	var jsonData map[string]any
	if err := json.Unmarshal(out.Bytes(), &jsonData); err != nil {
		t.Fatal(err)
	}

	data, ok := jsonData["data"].(map[string]any)
	if !ok {
		t.Fatal("data was not a map")
	}

	// Verify that the ID was sent
	val, exists := data["id"]
	if !exists {
		t.Fatal("Was expecting the data.id member to exist")
	}

	if val != id {
		t.Fatalf("Was expecting the data.id member to be `%s`, got `%s`", id, val)
	}
}

func TestMarshalOnePayload_omitIDString(t *testing.T) {
	type Foo struct {
		ID    string `jsonapi:"primary,foo"`
		Title string `jsonapi:"attr,title"`
	}

	foo := &Foo{Title: "Foo"}
	out := bytes.NewBuffer(nil)

	if err := MarshalPayload(out, foo); err != nil {
		t.Fatal(err)
	}

	var jsonData map[string]any
	if err := json.Unmarshal(out.Bytes(), &jsonData); err != nil {
		t.Fatal(err)
	}

	payload, ok := jsonData["data"].(map[string]any)
	if !ok {
		t.Fatal("data was not a map")
	}

	// Verify that empty ID of type string gets omitted. See:
	// https://github.com/google/jsonapi/issues/83#issuecomment-285611425
	_, ok = payload["id"]
	if ok {
		t.Fatal("Was expecting the data.id member to be omitted")
	}
}

func TestMarshall_invalidIDType(t *testing.T) {
	type badIDStruct struct {
		ID *bool `jsonapi:"primary,cars"`
	}

	id := true
	o := &badIDStruct{ID: &id}

	out := bytes.NewBuffer(nil)

	err := MarshalPayload(out, o)
	require.ErrorIs(t, err, ErrBadJSONAPIID)
}

func TestOmitsEmptyAnnotation(t *testing.T) {
	book := &Book{
		Author:      "aren55555",
		PublishedAt: time.Now().AddDate(0, -1, 0),
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, book); err != nil {
		t.Fatal(err)
	}

	var jsonData map[string]any
	if err := json.Unmarshal(out.Bytes(), &jsonData); err != nil {
		t.Fatal(err)
	}

	data, ok := jsonData["data"].(map[string]any)
	if !ok {
		t.Fatal("data was not a map")
	}

	attributes, ok := data["attributes"].(map[string]any)
	if !ok {
		t.Fatal("Was expecting the data.attributes key/value to have been a map")
	}

	// Verify that the specifically omitted field were omitted
	if val, exists := attributes["title"]; exists {
		t.Fatalf("Was expecting the data.attributes.title key/value to have been omitted - it was not and had a value of %v", val)
	}

	if val, exists := attributes["pages"]; exists {
		t.Fatalf("Was expecting the data.attributes.pages key/value to have been omitted - it was not and had a value of %v", val)
	}

	// Verify the implicitly omitted fields were omitted
	if val, exists := attributes["PublishedAt"]; exists {
		t.Fatalf("Was expecting the data.attributes.PublishedAt key/value to have been implicitly omitted - it was not and had a value of %v", val)
	}

	// Verify the unset fields were not omitted
	if _, exists := attributes["isbn"]; !exists {
		t.Fatal("Was expecting the data.attributes.isbn key/value to have NOT been omitted")
	}
}

func TestHasPrimaryAnnotation(t *testing.T) {
	testModel := &Blog{
		ID:        5,
		Title:     "Title 1",
		CreatedAt: time.Now(),
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, testModel); err != nil {
		t.Fatal(err)
	}

	resp := new(OnePayload)

	if err := json.NewDecoder(out).Decode(resp); err != nil {
		t.Fatal(err)
	}

	data := resp.Data

	if data.Type != "blogs" {
		t.Fatalf("type should have been blogs, got %s", data.Type)
	}

	if data.ID != "5" {
		t.Fatalf("ID not transferred")
	}
}

func TestSupportsAttributes(t *testing.T) {
	testModel := &Blog{
		ID:        5,
		Title:     "Title 1",
		CreatedAt: time.Now(),
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, testModel); err != nil {
		t.Fatal(err)
	}

	resp := new(OnePayload)
	if err := json.NewDecoder(out).Decode(resp); err != nil {
		t.Fatal(err)
	}

	data := resp.Data

	if data.Attributes == nil {
		t.Fatalf("Expected attributes")
	}

	if !bytes.Equal(data.Attributes["title"], json.RawMessage(`"Title 1"`)) {
		t.Fatalf("Attributes hash not populated using tags correctly")
	}
}

func TestOmitsZeroTimes(t *testing.T) {
	testModel := &Blog{
		ID:        5,
		Title:     "Title 1",
		CreatedAt: time.Time{},
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, testModel); err != nil {
		t.Fatal(err)
	}

	resp := new(OnePayload)
	if err := json.NewDecoder(out).Decode(resp); err != nil {
		t.Fatal(err)
	}

	data := resp.Data

	if data.Attributes == nil {
		t.Fatalf("Expected attributes")
	}

	if data.Attributes["created_at"] != nil {
		t.Fatalf("Created at was serialized even though it was a zero Time")
	}
}

func TestMarshalISO8601Time(t *testing.T) {
	testModel := &Timestamp{
		ID:   5,
		Time: time.Date(2016, 8, 17, 8, 27, 12, 23849, time.UTC),
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, testModel); err != nil {
		t.Fatal(err)
	}

	resp := new(OnePayload)
	if err := json.NewDecoder(out).Decode(resp); err != nil {
		t.Fatal(err)
	}

	data := resp.Data

	if data.Attributes == nil {
		t.Fatalf("Expected attributes")
	}

	if !bytes.Equal(data.Attributes["timestamp"], json.RawMessage(`"2016-08-17T08:27:12Z"`)) {
		t.Fatal("Timestamp was not serialised into ISO8601 correctly")
	}
}

func TestMarshalISO8601TimePointer(t *testing.T) {
	tm := time.Date(2016, 8, 17, 8, 27, 12, 23849, time.UTC)
	testModel := &Timestamp{
		ID:   5,
		Next: &tm,
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, testModel); err != nil {
		t.Fatal(err)
	}

	resp := new(OnePayload)
	if err := json.NewDecoder(out).Decode(resp); err != nil {
		t.Fatal(err)
	}

	data := resp.Data

	if data.Attributes == nil {
		t.Fatalf("Expected attributes")
	}

	if !bytes.Equal(data.Attributes["next"], json.RawMessage(`"2016-08-17T08:27:12Z"`)) {
		t.Fatalf("Next was not serialised into ISO8601 correctly got: %q", string(data.Attributes["next"]))
	}
}

func TestMarshalUnmarshalStructWithNestedFields(t *testing.T) {
	ts, err := isotime.ParseISO8601("2022-11-17T22:22:25.841137+01:00")
	require.NoError(t, err)

	s := &StructWithNestedFields{
		Timestamp:           ts,
		NestedStructPointer: &NestedField{NestedTimestamp: ts},
		NestedStruct:        NestedField{NestedTimestamp: ts},
	}

	in := new(bytes.Buffer)
	if err = MarshalPayload(in, s); err != nil {
		t.Fatal("Struct with nested fields errored while marshalling")
	}

	var out StructWithNestedFields
	if err = UnmarshalPayload(in, &out); err != nil {
		t.Fatal("Struct with nested fields errored while unmarshalling")
	}

	if !(out.Timestamp.Equal(out.NestedStructPointer.NestedTimestamp) &&
		out.Timestamp.Equal(out.NestedStruct.NestedTimestamp) &&
		out.NestedStruct.NestedTimestamp.Equal(out.NestedStructPointer.NestedTimestamp)) {
		t.Fatal("Time values in struct with nested fields are not equal")
	}
}

func TestMarshalDecimalPointer(t *testing.T) {
	dec := decimal.NewFromFloat(1.23)

	tcs := []struct {
		name string
		s    *StructWithNestedFields
	}{
		{
			name: "with non nil decimal pointer",
			s: &StructWithNestedFields{
				DecimalPointer: &dec,
			},
		},
		{
			name: "with nil decimal pointer",
			s: &StructWithNestedFields{
				DecimalPointer: nil,
			},
		},
		{
			name: "with nil nested decimal pointer with omit empty",
			s: &StructWithNestedFields{
				NestedStructPointer: &NestedField{
					DecimalPointerWithOmitEmpty: nil,
				},
			},
		},
		{
			name: "with nested zero decimal pointer with omit empty",
			s: &StructWithNestedFields{
				NestedStructPointer: &NestedField{
					DecimalPointerWithOmitEmpty: &decimal.Zero,
				},
			},
		},
		{
			name: "with nested non nil decimal pointer with omit empty",
			s: &StructWithNestedFields{
				NestedStructPointer: &NestedField{
					DecimalPointerWithOmitEmpty: &dec,
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			in := new(bytes.Buffer)
			if err := MarshalPayload(in, tc.s); err != nil {
				t.Fatal("Struct with nested fields errored while marshalling")
			}

			var out StructWithNestedFields
			if err := UnmarshalPayload(in, &out); err != nil {
				t.Fatal("Struct with nested fields errored while unmarshalling")
			}
		})
	}
}

func TestSupportsLinkable(t *testing.T) {
	testModel := &Blog{
		ID:        5,
		Title:     "Title 1",
		CreatedAt: time.Now(),
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, testModel); err != nil {
		t.Fatal(err)
	}

	resp := new(OnePayload)
	if err := json.NewDecoder(out).Decode(resp); err != nil {
		t.Fatal(err)
	}

	data := resp.Data

	if data.Links == nil {
		t.Fatal("Expected data.links")
	}

	links := *data.Links

	self, hasSelf := links["self"]
	if !hasSelf {
		t.Fatal("Expected 'self' link to be present")
	}

	if _, isString := self.(string); !isString {
		t.Fatal("Expected 'self' to contain a string")
	}

	comments, hasComments := links["comments"]
	if !hasComments {
		t.Fatal("expect 'comments' to be present")
	}

	commentsMap, isMap := comments.(map[string]any)
	if !isMap {
		t.Fatal("Expected 'comments' to contain a map")
	}

	commentsHref, hasHref := commentsMap["href"]
	if !hasHref {
		t.Fatal("Expect 'comments' to contain an 'href' key/value")
	}

	if _, isString := commentsHref.(string); !isString {
		t.Fatal("Expected 'href' to contain a string")
	}

	commentsMeta, hasMeta := commentsMap["meta"]
	if !hasMeta {
		t.Fatal("Expect 'comments' to contain a 'meta' key/value")
	}

	commentsMetaMap, isMap := commentsMeta.(map[string]any)
	if !isMap {
		t.Fatal("Expected 'comments' to contain a map")
	}

	commentsMetaObject := Meta(commentsMetaMap)

	countsMap, isMap := commentsMetaObject["counts"].(map[string]any)
	if !isMap {
		t.Fatal("Expected 'counts' to contain a map")
	}

	for k, v := range countsMap {
		if _, isNum := v.(float64); !isNum {
			t.Fatalf("Exepected value at '%s' to be a numeric (float64)", k)
		}
	}
}

func TestInvalidLinkable(t *testing.T) {
	testModel := &BadComment{
		ID:   5,
		Body: "Hello World",
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, testModel); err == nil {
		t.Fatal("Was expecting an error")
	}
}

func TestSupportsMetable(t *testing.T) {
	testModel := &Blog{
		ID:        5,
		Title:     "Title 1",
		CreatedAt: time.Now(),
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, testModel); err != nil {
		t.Fatal(err)
	}

	resp := new(OnePayload)
	if err := json.NewDecoder(out).Decode(resp); err != nil {
		t.Fatal(err)
	}

	data := resp.Data
	if data.Meta == nil {
		t.Fatalf("Expected data.meta")
	}

	meta := *data.Meta
	if e, a := "extra details regarding the blog", meta["detail"]; e != a {
		t.Fatalf("Was expecting meta.detail to be %q, got %q", e, a)
	}
}

func TestRelations(t *testing.T) {
	testModel := testBlog()

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, testModel); err != nil {
		t.Fatal(err)
	}

	resp := new(OnePayload)
	if err := json.NewDecoder(out).Decode(resp); err != nil {
		t.Fatal(err)
	}

	relations := resp.Data.Relationships

	if relations == nil {
		t.Fatalf("Relationships were not materialized")
	}

	if relations["posts"] == nil {
		t.Fatalf("Posts relationship was not materialized")
	} else {
		if posts, ok := relations["posts"].(map[string]any); !ok || posts["links"] == nil {
			t.Fatalf("Posts relationship links were not materialized")
		}

		if posts, ok := relations["posts"].(map[string]any); !ok || posts["meta"] == nil {
			t.Fatalf("Posts relationship meta were not materialized")
		}
	}

	if relations["current_post"] == nil {
		t.Fatalf("Current post relationship was not materialized")
	} else {
		if currentPost, ok := relations["current_post"].(map[string]any); !ok || currentPost["links"] == nil {
			t.Fatalf("Current post relationship links were not materialized")
		}

		if currentPost, ok := relations["current_post"].(map[string]any); !ok || currentPost["meta"] == nil {
			t.Fatalf("Current post relationship meta were not materialized")
		}
	}

	posts, ok := relations["posts"].(map[string]any)
	if !ok {
		t.Fatalf("Expected posts to be a map")
	}

	postsData, ok := posts["data"].([]any)
	if !ok {
		t.Fatalf("Expected posts.data to be a slice")
	}

	if len(postsData) != 2 {
		t.Fatalf("Did not materialize two posts")
	}
}

func TestNoRelations(t *testing.T) {
	testModel := &Blog{ID: 1, Title: "Title 1", CreatedAt: time.Now()}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, testModel); err != nil {
		t.Fatal(err)
	}

	resp := new(OnePayload)
	if err := json.NewDecoder(out).Decode(resp); err != nil {
		t.Fatal(err)
	}

	if resp.Included != nil {
		t.Fatalf("Encoding json response did not omit included")
	}
}

func TestMarshalPayloadWithoutIncluded(t *testing.T) {
	data := &Post{
		ID:       1,
		BlogID:   2,
		ClientID: "123e4567-e89b-12d3-a456-426655440000",
		Title:    "Foo",
		Body:     "Bar",
		Comments: []*Comment{
			{
				ID:   20,
				Body: "First",
			},
			{
				ID:   21,
				Body: "Hello World",
			},
		},
		LatestComment: &Comment{
			ID:   22,
			Body: "Cool!",
		},
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayloadWithoutIncluded(out, data); err != nil {
		t.Fatal(err)
	}

	resp := new(OnePayload)
	if err := json.NewDecoder(out).Decode(resp); err != nil {
		t.Fatal(err)
	}

	if resp.Included != nil {
		t.Fatalf("Encoding json response did not omit included")
	}
}

func TestMarshalPayload_many(t *testing.T) {
	data := []any{
		&Blog{
			ID:        5,
			Title:     "Title 1",
			CreatedAt: time.Now(),
			Posts: []*Post{
				{
					ID:    1,
					Title: "Foo",
					Body:  "Bar",
				},
				{
					ID:    2,
					Title: "Fuubar",
					Body:  "Bas",
				},
			},
			CurrentPost: &Post{
				ID:    1,
				Title: "Foo",
				Body:  "Bar",
			},
		},
		&Blog{
			ID:        6,
			Title:     "Title 2",
			CreatedAt: time.Now(),
			Posts: []*Post{
				{
					ID:    3,
					Title: "Foo",
					Body:  "Bar",
				},
				{
					ID:    4,
					Title: "Fuubar",
					Body:  "Bas",
				},
			},
			CurrentPost: &Post{
				ID:    4,
				Title: "Foo",
				Body:  "Bar",
			},
		},
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, data); err != nil {
		t.Fatal(err)
	}

	resp := new(ManyPayload)
	if err := json.NewDecoder(out).Decode(resp); err != nil {
		t.Fatal(err)
	}

	d := resp.Data

	if len(d) != 2 {
		t.Fatalf("data should have two elements")
	}
}

func TestMarshalMany_WithSliceOfStructPointers(t *testing.T) {
	var data []*Blog
	for len(data) < 2 {
		data = append(data, testBlog())
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayload(out, data); err != nil {
		t.Fatal(err)
	}

	resp := new(ManyPayload)
	if err := json.NewDecoder(out).Decode(resp); err != nil {
		t.Fatal(err)
	}

	d := resp.Data

	if len(d) != 2 {
		t.Fatalf("data should have two elements")
	}
}

func TestMarshalManyWithoutIncluded(t *testing.T) {
	var data []*Blog
	for len(data) < 2 {
		data = append(data, testBlog())
	}

	out := bytes.NewBuffer(nil)
	if err := MarshalPayloadWithoutIncluded(out, data); err != nil {
		t.Fatal(err)
	}

	resp := new(ManyPayload)
	if err := json.NewDecoder(out).Decode(resp); err != nil {
		t.Fatal(err)
	}

	d := resp.Data

	if len(d) != 2 {
		t.Fatalf("data should have two elements")
	}

	if resp.Included != nil {
		t.Fatalf("Encoding json response did not omit included")
	}
}

func TestMarshalMany_SliceOfInterfaceAndSliceOfStructsSameJSON(t *testing.T) {
	structs := []*Book{
		{ID: 1, Author: "aren55555", ISBN: "abc"},
		{ID: 2, Author: "shwoodard", ISBN: "xyz"},
	}
	interfaces := []any{}

	for _, s := range structs {
		interfaces = append(interfaces, s)
	}

	// Perform Marshals
	structsOut := new(bytes.Buffer)
	if err := MarshalPayload(structsOut, structs); err != nil {
		t.Fatal(err)
	}

	interfacesOut := new(bytes.Buffer)
	if err := MarshalPayload(interfacesOut, interfaces); err != nil {
		t.Fatal(err)
	}

	// Generic JSON Unmarshal
	structsData, interfacesData := make(map[string]any), make(map[string]any)
	if err := json.Unmarshal(structsOut.Bytes(), &structsData); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(interfacesOut.Bytes(), &interfacesData); err != nil {
		t.Fatal(err)
	}

	// Compare Result
	if !reflect.DeepEqual(structsData, interfacesData) {
		t.Fatal("Was expecting the JSON API generated to be the same")
	}
}

func TestMarshal_InvalidIntefaceArgument(t *testing.T) {
	out := new(bytes.Buffer)

	err := MarshalPayload(out, true)
	require.ErrorIs(t, err, ErrUnexpectedType)

	err = MarshalPayload(out, 25)
	require.ErrorIs(t, err, ErrUnexpectedType)

	err = MarshalPayload(out, Book{})
	require.ErrorIs(t, err, ErrUnexpectedType)
}

func testBlog() *Blog {
	return &Blog{
		ID:        5,
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
