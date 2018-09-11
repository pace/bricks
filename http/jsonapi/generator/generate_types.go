// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package generator

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
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
		// create new type
		name := goNameHelper(name)

		// skip jsonapi error type
		if name == "Errors" {
			continue
		}

		t := jen.Type().Id(name)
		err := g.buildType(name, t, schemaType)
		if err != nil {
			return err
		}
		// document type
		g.addGoDoc(name, schemaType.Value.Description)
		g.goSource.Add(t)
	}

	return nil
}

func (g *Generator) buildType(prefix string, stmt *jen.Statement, schema *openapi3.SchemaRef) error {
	// handle references
	if schema.Ref != "" {
		// if there is a reference to a type use it
		stmt.Op("*").Id(goNameHelper(filepath.Base(schema.Ref)))
		return nil
	}

	val := schema.Value

	switch val.Type {
	case "array":
		return g.buildType(prefix, stmt.Index(), val.Items)
	case "object":
		if data := val.Properties["data"]; data != nil {
			if data.Ref != "" {
				return g.buildType(prefix+"Ref", stmt, data)
			} else if data.Value.Type == "array" {
				item := prefix + "Item"
				stmt.Index().Op("*").Id(item)
				g.addGoDoc(item, data.Value.Description)
				itemStmt := g.goSource.Type().Id(item)
				return g.structJSONAPI(prefix, itemStmt, data.Value.Items.Value)
			}

			return g.structJSONAPI(prefix, stmt, data.Value)
		} else if val.Properties["id"] != nil &&
			val.Properties["type"] != nil &&
			(val.Properties["attributes"] != nil ||
				val.Properties["relationships"] != nil) {
			return g.structJSONAPI(prefix, stmt, val)
		}

		return g.buildTypeStruct(prefix, stmt, val)
	default:
		// skip allOf, anyOf and oneOf, as they can't be generated
		if len(val.AllOf)+len(val.AnyOf)+len(val.OneOf) > 0 {
			log.Warnf("Can't generate allOf, anyOf and oneOf for type %q", prefix)
			stmt.Qual("encoding/json", "RawMessage")
			return nil
		}

		err := g.goType(stmt, val, make(map[string]string))
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) buildTypeStruct(name string, stmt *jen.Statement, schema *openapi3.Schema) error {
	// build regular struct
	fields, err := g.generateStructFields(name, schema, false)
	if err != nil {
		return err
	}

	// generate struct as separate objects to allow direct creation of those objects
	g.addGoDoc(name, schema.Description)
	g.goSource.Type().Id(name).Struct(fields...)

	// use new struct pointer
	stmt.Op("*").Id(name)
	return nil
}

// references the type from the schema or generates a new type (inline)
// and returns
func (g *Generator) generateTypeReference(fallbackName string, schema *openapi3.SchemaRef) (jen.Code, error) {
	// handle references
	if schema.Ref != "" {
		// if there is a reference to a type use it
		return jen.Id(goNameHelper(filepath.Base(schema.Ref))), nil
	}

	// generate type and doc as a fallback (if no ref provided)
	g.addGoDoc(fallbackName, schema.Value.Description)
	err := g.buildType(fallbackName, g.goSource.Type().Id(fallbackName), schema)
	if err != nil {
		return nil, err
	}

	return jen.Op("*").Id(fallbackName), nil
}

func (g *Generator) structJSONAPI(prefix string, stmt *jen.Statement, schema *openapi3.Schema) error {
	var fields []jen.Code

	propID := schema.Properties["id"]
	propType := schema.Properties["type"]

	if propID == nil || propType == nil {
		return fmt.Errorf("ID/Type missing for jsonapi type %q", prefix)
	}

	// add ID
	id, err := g.generateIDField(propID.Value, propType.Value)
	if err != nil {
		return err
	}
	fields = append(fields, id)

	// add attributes
	if attr := schema.Properties["attributes"]; attr != nil {
		attrFields, err := g.generateStructFields(prefix, attr.Value, true)
		if err != nil {
			return err
		}
		fields = append(fields, attrFields...)
	}

	// add relationships
	if rels := schema.Properties["relationships"]; rels != nil {
		relFields, err := g.generateStructRelationships(prefix, rels.Value, true)
		if err != nil {
			return err
		}
		fields = append(fields, relFields...)
	}

	stmt.Struct(fields...)
	return nil
}

func (g *Generator) generateAttrField(prefix, name string, schema *openapi3.SchemaRef, tags map[string]string) (*jen.Statement, error) {
	field := jen.Id(goNameHelper(name))

	err := g.buildType(prefix+goNameHelper(name), field, schema)
	if err != nil {
		return nil, err
	}
	field.Tag(tags)
	if schema.Ref == "" {
		g.commentOrExample(field, schema.Value)
	}
	return field, nil
}

func (g *Generator) generateStructFields(prefix string, schema *openapi3.Schema, jsonAPIObject bool) ([]jen.Code, error) {
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
		addJSONAPITags(tags, attrName, jsonAPIObject)
		addRequiredOptionalTag(tags, attrName, schema)

		// generate attribute field
		field, err := g.generateAttrField(prefix, attrName, attrSchema, tags)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func (g *Generator) generateStructRelationships(prefix string, schema *openapi3.Schema, jsonAPI bool) ([]jen.Code, error) {
	// sort by key
	keys := make([]string, 0, len(schema.Properties))
	for k := range schema.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var relationships []jen.Code
	for _, relName := range keys {
		relSchema := schema.Properties[relName]
		tags := make(map[string]string)
		addJSONAPITags(tags, relName, jsonAPI)
		addRequiredOptionalTag(tags, relName, schema)

		// check for data
		data := relSchema.Value.Properties["data"]
		if data == nil || data.Value == nil {
			return nil, fmt.Errorf("No data for relationship %s context %s", relName, prefix)
		}

		// generate relationship field
		rel := jen.Id(goNameHelper(relName))

		switch data.Value.Type {
		case "array": // one-to-many
			name := data.Value.Items.Value.Properties["type"].Value.Enum[0].(string)
			rel.Index().Op("*").Id(goNameHelper(name)).Tag(tags)
		case "object": // belongs-to
			name := data.Value.Properties["type"].Value.Enum[0].(string)
			rel.Op("*").Id(goNameHelper(name)).Tag(tags)
		}

		relationships = append(relationships, rel)
	}
	return relationships, nil
}

func (g *Generator) generateIDField(idType, objectType *openapi3.Schema) (*jen.Statement, error) {
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
	g.commentOrExample(id, idType)
	return id, nil
}

func addRequiredOptionalTag(tags map[string]string, name string, schema *openapi3.Schema) {
	// check if field is required
	isRequired := false
	for _, required := range schema.Required {
		if required == name {
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
}

func addJSONAPITags(tags map[string]string, name string, jsonAPI bool) {
	// Add json-api tag
	if jsonAPI {
		tags["jsonapi"] = fmt.Sprintf("attr,%s,omitempty", name)
	} else {
		tags["jsonapi"] = fmt.Sprintf("%s,omitempty", name)
	}
	// Add json tag
	tags["json"] = fmt.Sprintf("%s,omitempty", name)
}
