// This file is originating from https://github.com/google/jsonapi/
// To this file the license conditions of LICENSE apply.

package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var (
	// ErrBadJSONAPIStructTag is returned when the Struct field's JSON API
	// annotation is invalid.
	ErrBadJSONAPIStructTag = errors.New("bad jsonapi struct tag format")
	// ErrBadJSONAPIID is returned when the Struct JSON API annotated "id" field
	// was not a valid numeric type.
	ErrBadJSONAPIID = errors.New(
		"id should be either string, int(8,16,32,64) or uint(8,16,32,64)")
	// ErrExpectedSlice is returned when a variable or argument was expected to
	// be a slice of *Structs; MarshalMany will return this error when its
	// any argument is invalid.
	ErrExpectedSlice = errors.New("models should be a slice of struct pointers")
	// ErrUnexpectedType is returned when marshalling an interface; the interface
	// had to be a pointer or a slice; otherwise this error is returned.
	ErrUnexpectedType = errors.New("models should be a struct pointer or slice of struct pointers")
)

// MarshalPayload writes a jsonapi response for one or many records. The
// related records are sideloaded into the "included" array. If this method is
// given a struct pointer as an argument it will serialize in the form
// "data": {...}. If this method is given a slice of pointers, this method will
// serialize in the form "data": [...]
//
// One Example: you could pass it, w, your http.ResponseWriter, and, models, a
// ptr to a Blog to be written to the response body:
//
//	 func ShowBlog(w http.ResponseWriter, r *http.Request) {
//		 blog := &Blog{}
//
//		 w.Header().Set("Content-Type", jsonapi.MediaType)
//		 w.WriteHeader(http.StatusOK)
//
//		 if err := jsonapi.MarshalPayload(w, blog); err != nil {
//			 http.Error(w, err.Error(), http.StatusInternalServerError)
//		 }
//	 }
//
// Many Example: you could pass it, w, your http.ResponseWriter, and, models, a
// slice of Blog struct instance pointers to be written to the response body:
//
//		 func ListBlogs(w http.ResponseWriter, r *http.Request) {
//	    blogs := []*Blog{}
//
//			 w.Header().Set("Content-Type", jsonapi.MediaType)
//			 w.WriteHeader(http.StatusOK)
//
//			 if err := jsonapi.MarshalPayload(w, blogs); err != nil {
//				 http.Error(w, err.Error(), http.StatusInternalServerError)
//			 }
//		 }
func MarshalPayload(w io.Writer, models any) error {
	payload, err := Marshal(models)
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(payload)
}

// Marshal does the same as MarshalPayload except it just returns the payload
// and doesn't write out results. Useful if you use your own JSON rendering
// library.
func Marshal(models any) (Payloader, error) {
	switch vals := reflect.ValueOf(models); vals.Kind() {
	case reflect.Slice:
		m, err := convertToSliceInterface(&models)
		if err != nil {
			return nil, err
		}

		payload, err := marshalMany(m)
		if err != nil {
			return nil, err
		}

		if linkableModels, isLinkable := models.(Linkable); isLinkable {
			jl := linkableModels.JSONAPILinks()
			if er := jl.validate(); er != nil {
				return nil, er
			}

			payload.Links = linkableModels.JSONAPILinks()
		}

		if metableModels, ok := models.(Metable); ok {
			payload.Meta = metableModels.JSONAPIMeta()
		}

		return payload, nil
	case reflect.Ptr:
		// Check that the pointer was to a struct
		if reflect.Indirect(vals).Kind() != reflect.Struct {
			return nil, ErrUnexpectedType
		}

		return marshalOne(models)
	default:
		return nil, ErrUnexpectedType
	}
}

// MarshalPayloadWithoutIncluded writes a jsonapi response with one or many
// records, without the related records sideloaded into "included" array.
// If you want to serialize the relations into the "included" array see
// MarshalPayload.
//
// models any should be either a struct pointer or a slice of struct
// pointers.
func MarshalPayloadWithoutIncluded(w io.Writer, model any) error {
	payload, err := Marshal(model)
	if err != nil {
		return err
	}

	payload.clearIncluded()

	return json.NewEncoder(w).Encode(payload)
}

// marshalOne does the same as MarshalOnePayload except it just returns the
// payload and doesn't write out results. Useful is you use your JSON rendering
// library.
func marshalOne(model any) (*OnePayload, error) {
	included := make(map[string]*Node)

	rootNode, err := visitModelNode(model, &included, true)
	if err != nil {
		return nil, err
	}

	payload := &OnePayload{Data: rootNode}

	payload.Included = nodeMapValues(&included)

	return payload, nil
}

// marshalMany does the same as MarshalManyPayload except it just returns the
// payload and doesn't write out results. Useful is you use your JSON rendering
// library.
func marshalMany(models []any) (*ManyPayload, error) {
	payload := &ManyPayload{
		Data: []*Node{},
	}
	included := map[string]*Node{}

	for _, model := range models {
		node, err := visitModelNode(model, &included, true)
		if err != nil {
			return nil, err
		}

		payload.Data = append(payload.Data, node)
	}

	payload.Included = nodeMapValues(&included)

	return payload, nil
}

// MarshalOnePayloadEmbedded - This method not meant to for use in
// implementation code, although feel free.  The purpose of this
// method is for use in tests.  In most cases, your request
// payloads for create will be embedded rather than sideloaded for
// related records. This method will serialize a single struct
// pointer into an embedded json response. In other words, there
// will be no, "included", array in the json all relationships will
// be serailized inline in the data.
//
// However, in tests, you may want to construct payloads to post
// to create methods that are embedded to most closely resemble
// the payloads that will be produced by the client. This is what
// this method is intended for.
//
// model any should be a pointer to a struct.
func MarshalOnePayloadEmbedded(w io.Writer, model any) error {
	rootNode, err := visitModelNode(model, nil, false)
	if err != nil {
		return err
	}

	payload := &OnePayload{Data: rootNode}

	return json.NewEncoder(w).Encode(payload)
}

func visitModelNode(model any, included *map[string]*Node,
	sideload bool,
) (*Node, error) {
	node := new(Node)

	var er error

	value := reflect.ValueOf(model)
	if value.IsNil() {
		return nil, nil
	}

	modelValue := value.Elem()
	modelType := value.Type().Elem()

	for i := range modelValue.NumField() {
		structField := modelValue.Type().Field(i)

		tag := structField.Tag.Get(annotationJSONAPI)
		if tag == "" {
			continue
		}

		fieldValue := modelValue.Field(i)
		fieldType := modelType.Field(i)

		args := strings.Split(tag, annotationSeperator)

		if len(args) < 1 {
			er = ErrBadJSONAPIStructTag
			break
		}

		annotation := args[0]

		if (annotation == annotationClientID && len(args) != 1) ||
			(annotation != annotationClientID && len(args) < 2) {
			er = ErrBadJSONAPIStructTag
			break
		}

		if annotation == annotationPrimary {
			v := fieldValue

			// Deal with PTRS
			var kind reflect.Kind
			if fieldValue.Kind() == reflect.Ptr {
				kind = fieldType.Type.Elem().Kind()
				v = reflect.Indirect(fieldValue)
			} else {
				kind = fieldType.Type.Kind()
			}

			// Handle allowed types
			switch kind {
			case reflect.String:
				node.ID, _ = v.Interface().(string)
			case reflect.Int:
				val, ok := v.Interface().(int)
				if !ok {
					return nil, errors.New("could not assert int")
				}

				node.ID = strconv.FormatInt(int64(val), 10)
			case reflect.Int8:
				val, ok := v.Interface().(int8)
				if !ok {
					return nil, errors.New("could not assert int8")
				}

				node.ID = strconv.FormatInt(int64(val), 10)
			case reflect.Int16:
				val, ok := v.Interface().(int16)
				if !ok {
					return nil, errors.New("could not assert int16")
				}

				node.ID = strconv.FormatInt(int64(val), 10)
			case reflect.Int32:
				val, ok := v.Interface().(int32)
				if !ok {
					return nil, errors.New("could not assert int32")
				}

				node.ID = strconv.FormatInt(int64(val), 10)
			case reflect.Int64:
				val, ok := v.Interface().(int64)
				if !ok {
					return nil, errors.New("could not assert int64")
				}

				node.ID = strconv.FormatInt(val, 10)
			case reflect.Uint:
				val, ok := v.Interface().(uint)
				if !ok {
					return nil, errors.New("could not assert uint")
				}

				node.ID = strconv.FormatUint(uint64(val), 10)
			case reflect.Uint8:
				val, ok := v.Interface().(uint8)
				if !ok {
					return nil, errors.New("could not assert uint8")
				}

				node.ID = strconv.FormatUint(uint64(val), 10)
			case reflect.Uint16:
				val, ok := v.Interface().(uint16)
				if !ok {
					return nil, errors.New("could not assert uint16")
				}

				node.ID = strconv.FormatUint(uint64(val), 10)
			case reflect.Uint32:
				val, ok := v.Interface().(uint32)
				if !ok {
					return nil, errors.New("could not assert uint32")
				}

				node.ID = strconv.FormatUint(uint64(val), 10)
			case reflect.Uint64:
				val, ok := v.Interface().(uint64)
				if !ok {
					return nil, errors.New("could not assert uint64")
				}

				node.ID = strconv.FormatUint(val, 10)
			default:
				// We had a JSON float (numeric), but our field was not one of the
				// allowed numeric types
				er = ErrBadJSONAPIID
			}

			if er != nil {
				break
			}

			node.Type = args[1]
		} else if annotation == annotationClientID {
			clientID := fieldValue.String()
			if clientID != "" {
				node.ClientID = clientID
			}
		} else if annotation == annotationAttribute {
			var omitEmpty, iso8601 bool

			if len(args) > 2 {
				for _, arg := range args[2:] {
					switch arg {
					case annotationOmitEmpty:
						omitEmpty = true
					case annotationISO8601:
						iso8601 = true
					}
				}
			}

			if node.Attributes == nil {
				node.Attributes = make(map[string]json.RawMessage)
			}

			var err error

			if fieldValue.Type() == reflect.TypeOf(decimal.Decimal{}) {
				d, ok := fieldValue.Interface().(decimal.Decimal)
				if !ok {
					return nil, fmt.Errorf("could not assert decimal.Decimal")
				}

				if !decimal.MarshalJSONWithoutQuotes {
					return nil, fmt.Errorf("decimal.MarshalJSONWithoutQuotes needs to be turned on to export decimals as numbers")
				}

				node.Attributes[args[1]] = json.RawMessage(d.String())
			} else if fieldValue.Type() == reflect.TypeOf(new(decimal.Decimal)) {
				// A decimal pointer may be nil
				if fieldValue.IsNil() {
					if omitEmpty {
						continue
					}

					node.Attributes[args[1]] = []byte("null")
				} else {
					d, ok := fieldValue.Interface().(*decimal.Decimal)
					if !ok {
						return nil, fmt.Errorf("could not assert decimal.Decimal")
					}

					if !decimal.MarshalJSONWithoutQuotes {
						return nil, fmt.Errorf("decimal.MarshalJSONWithoutQuotes needs to be turned on to export decimals as numbers")
					}

					node.Attributes[args[1]] = json.RawMessage(d.String())
				}
			} else if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
				t, ok := fieldValue.Interface().(time.Time)
				if !ok {
					return nil, fmt.Errorf("could not assert time.Time")
				}

				if t.IsZero() {
					continue
				}

				if iso8601 {
					node.Attributes[args[1]], err = json.Marshal(t.UTC().Format(iso8601TimeFormat))
				} else {
					node.Attributes[args[1]], err = json.Marshal(t.Unix())
				}

				if err != nil {
					return nil, err
				}
			} else if fieldValue.Type() == reflect.TypeOf(new(time.Time)) {
				// A time pointer may be nil
				if fieldValue.IsNil() {
					if omitEmpty {
						continue
					}

					node.Attributes[args[1]] = []byte("null")
				} else {
					tm, ok := fieldValue.Interface().(*time.Time)
					if !ok {
						return nil, fmt.Errorf("could not assert time.Time")
					}

					if tm.IsZero() && omitEmpty {
						continue
					}

					if iso8601 {
						node.Attributes[args[1]], err = json.Marshal(tm.UTC().Format(iso8601TimeFormat))
					} else {
						node.Attributes[args[1]], err = json.Marshal(tm.Unix())
					}

					if err != nil {
						return nil, err
					}
				}
			} else {
				// Dealing with a fieldValue that is not a time
				emptyValue := reflect.Zero(fieldValue.Type())

				// See if we need to omit this field
				if omitEmpty && reflect.DeepEqual(fieldValue.Interface(), emptyValue.Interface()) {
					continue
				}

				if strAttr, ok := fieldValue.Interface().(string); ok {
					node.Attributes[args[1]], err = json.Marshal(strAttr)
				} else if fieldValue.Type().Kind() == reflect.Struct {
					// We need to pass a pointer value
					ptr := reflect.New(fieldValue.Type())
					ptr.Elem().Set(fieldValue)

					n, err1 := visitModelNode(ptr.Interface(), nil, false)
					if err1 != nil {
						return nil, err1
					}

					node.Attributes[args[1]], err = json.Marshal(n.Attributes)
				} else if fieldValue.Type().Kind() == reflect.Ptr && fieldValue.Elem().Kind() == reflect.Struct {
					n, err1 := visitModelNode(fieldValue.Interface(), nil, false)
					if err1 != nil {
						return nil, err1
					}

					node.Attributes[args[1]], err = json.Marshal(n.Attributes)
				} else {
					node.Attributes[args[1]], err = json.Marshal(fieldValue.Interface())
				}

				if err != nil {
					return nil, err
				}
			}
		} else if annotation == annotationRelation {
			var omitEmpty bool

			// add support for 'omitempty' struct tag for marshaling as absent
			if len(args) > 2 {
				omitEmpty = args[2] == annotationOmitEmpty
			}

			isSlice := fieldValue.Type().Kind() == reflect.Slice
			if omitEmpty &&
				(isSlice && fieldValue.Len() < 1 ||
					(!isSlice && fieldValue.IsNil())) {
				continue
			}

			if node.Relationships == nil {
				node.Relationships = make(map[string]any)
			}

			var relLinks *Links
			if linkableModel, ok := model.(RelationshipLinkable); ok {
				relLinks = linkableModel.JSONAPIRelationshipLinks(args[1])
			}

			var relMeta *Meta
			if metableModel, ok := model.(RelationshipMetable); ok {
				relMeta = metableModel.JSONAPIRelationshipMeta(args[1])
			}

			if isSlice {
				// to-many relationship
				relationship, err := visitModelNodeRelationships(
					fieldValue,
					included,
					sideload,
				)
				if err != nil {
					er = err
					break
				}

				relationship.Links = relLinks
				relationship.Meta = relMeta

				if sideload {
					shallowNodes := []*Node{}

					for _, n := range relationship.Data {
						appendIncluded(included, n)
						shallowNodes = append(shallowNodes, toShallowNode(n))
					}

					node.Relationships[args[1]] = &RelationshipManyNode{
						Data:  shallowNodes,
						Links: relationship.Links,
						Meta:  relationship.Meta,
					}
				} else {
					node.Relationships[args[1]] = relationship
				}
			} else {
				// to-one relationships
				// Handle null relationship case
				if fieldValue.IsNil() {
					node.Relationships[args[1]] = &RelationshipOneNode{Data: nil}
					continue
				}

				relationship, err := visitModelNode(
					fieldValue.Interface(),
					included,
					sideload,
				)
				if err != nil {
					er = err
					break
				}

				if sideload {
					appendIncluded(included, relationship)

					node.Relationships[args[1]] = &RelationshipOneNode{
						Data:  toShallowNode(relationship),
						Links: relLinks,
						Meta:  relMeta,
					}
				} else {
					node.Relationships[args[1]] = &RelationshipOneNode{
						Data:  relationship,
						Links: relLinks,
						Meta:  relMeta,
					}
				}
			}
		} else {
			er = ErrBadJSONAPIStructTag
			break
		}
	}

	if er != nil {
		return nil, er
	}

	if linkableModel, isLinkable := model.(Linkable); isLinkable {
		jl := linkableModel.JSONAPILinks()
		if er := jl.validate(); er != nil {
			return nil, er
		}

		node.Links = linkableModel.JSONAPILinks()
	}

	if metableModel, ok := model.(Metable); ok {
		node.Meta = metableModel.JSONAPIMeta()
	}

	return node, nil
}

func toShallowNode(node *Node) *Node {
	return &Node{
		ID:   node.ID,
		Type: node.Type,
	}
}

func visitModelNodeRelationships(models reflect.Value, included *map[string]*Node,
	sideload bool,
) (*RelationshipManyNode, error) {
	nodes := []*Node{}

	for i := range models.Len() {
		n := models.Index(i).Interface()

		node, err := visitModelNode(n, included, sideload)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
	}

	return &RelationshipManyNode{Data: nodes}, nil
}

func appendIncluded(m *map[string]*Node, nodes ...*Node) {
	included := *m

	for _, n := range nodes {
		k := fmt.Sprintf("%s,%s", n.Type, n.ID)

		if _, hasNode := included[k]; hasNode {
			continue
		}

		included[k] = n
	}
}

func nodeMapValues(m *map[string]*Node) []*Node {
	mp := *m
	nodes := make([]*Node, len(mp))

	i := 0

	for _, n := range mp {
		nodes[i] = n
		i++
	}

	return nodes
}

func convertToSliceInterface(i *any) ([]any, error) {
	vals := reflect.ValueOf(*i)
	if vals.Kind() != reflect.Slice {
		return nil, ErrExpectedSlice
	}

	response := make([]any, 0, vals.Len())

	for x := range vals.Len() {
		response = append(response, vals.Index(x).Interface())
	}

	return response, nil
}
