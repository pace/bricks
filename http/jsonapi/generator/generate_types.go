// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
)

// BuildTypes transforms all component schemas into go types
func (g *Generator) BuildTypes(schema *openapi3.Swagger) error {
	schemas := schema.Components.Schemas

	// sort by key
	keys := make([]string, 0, len(schemas))
	for k := range schemas {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		schemaType := schemas[name]
		err := g.buildType(name, schemaType.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) buildType(name string, schema *openapi3.Schema) error {
	// document type
	g.addGoDoc(name, schema.Description)

	// create new type
	t := g.goSource.Type().Id(name)

	// extract data
	if data := schema.Properties["data"]; data != nil && data.Value != nil {
		// jsonapi object
		switch data.Value.Type {
		case "array":
			err := g.generateArray(name, t, data.Value)
			if err != nil {
				return err
			}
		case "object":
			err := g.structJsonApi(name, t, data.Value)
			if err != nil {
				return err
			}
		}
	} else {
		fields, err := g.generateStructFields(name, schema, false)
		if err != nil {
			return err
		}
		t.Struct(fields...)
	}

	return nil
}

func (g *Generator) generateArray(prefix string, stmt *jen.Statement, schema *openapi3.Schema) error {
	// geneate regular arrays of jsonapi structs if id and type are present
	if schema.Items.Value.Properties["id"] != nil && schema.Items.Value.Properties["type"] != nil {
		return g.structJsonApi(prefix, stmt.Index(), schema.Items.Value)
	}

	// generate embedded struct
	switch schema.Items.Value.Type {
	case "object":
		err := g.generateEmbedStruct(prefix, schema.Items.Value)
		if err != nil {
			return err
		}
		stmt.Index().Id("*" + prefix)
	default:
		err := g.goType(stmt.Index(), schema.Items.Value, make(map[string]string))
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) structJsonApi(prefix string, stmt *jen.Statement, schema *openapi3.Schema) error {
	var fields []jen.Code

	// add ID
	id, err := g.generateIdField(schema.Properties["id"].Value, schema.Properties["type"].Value)
	if err != nil {
		return err
	}
	fields = append(fields, id)

	// add attributes
	attrFields, err := g.generateStructFields(prefix, schema.Properties["attributes"].Value, true)
	if err != nil {
		return err
	}
	fields = append(fields, attrFields...)

	stmt.Struct(fields...)
	return nil
}

func (g *Generator) generateAttrField(prefix, name string, schema *openapi3.Schema, jsonApi bool, tags map[string]string) (*jen.Statement, error) {
	field := jen.Id(goNameHelper(name))

	// Add json-api tag
	if jsonApi {
		tags["jsonapi"] = fmt.Sprintf("attr,%s,omitempty", name)
	} else {
		tags["jsonapi"] = fmt.Sprintf("%s,omitempty", name)
	}

	switch schema.Type {
	case "array":
		g.generateArray(prefix+strings.Title(name), field, schema)
	case "object": // embedded json object are not json-api
		prefix := prefix + strings.Title(name)
		err := g.generateEmbedStruct(prefix, schema)
		if err != nil {
			return nil, err
		}
		field.Id("*" + prefix)
	default:
		err := g.goType(field, schema, tags)
		if err != nil {
			return nil, err
		}
	}
	field.Tag(tags)
	g.CommentOrExample(field, schema)
	return field, nil
}

func (g *Generator) generateEmbedStruct(prefix string, schema *openapi3.Schema) error {
	// document type
	g.addGoDoc(prefix, schema.Description)
	embed := g.goSource.Type().Id(prefix)

	fields, err := g.generateStructFields(prefix, schema, false)
	if err != nil {
		return err
	}
	embed.Struct(fields...)
	return nil
}

func (g *Generator) generateStructFields(prefix string, schema *openapi3.Schema, jsonApi bool) ([]jen.Code, error) {
	// sort by key
	keys := make([]string, 0, len(schema.Properties))
	for k := range schema.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var fields []jen.Code
	for _, attrName := range keys {
		attrSchema := schema.Properties[attrName]
		tags := make(map[string]string)

		// check if field is required
		isRequired := false
		for _, required := range schema.Required {
			if required == attrName {
				isRequired = true
				break
			}
		}

		// add required if otherwise optional validation
		if isRequired {
			addValidator(tags, "required")
		} else {
			addValidator(tags, "optional")
		}

		// generate attribute field
		field, err := g.generateAttrField(prefix, attrName, attrSchema.Value, jsonApi, tags)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func (g *Generator) generateIdField(idType, objectType *openapi3.Schema) (*jen.Statement, error) {
	id := jen.Id("ID")
	tags := map[string]string{
		"jsonapi": fmt.Sprintf("primary,%s,omitempty", objectType.Enum[0]),
	}
	err := g.goType(id, idType, tags)
	if err != nil {
		return nil, err
	}
	addValidator(tags, "optional")
	id.Tag(tags)
	g.CommentOrExample(id, idType)
	return id, nil
}
