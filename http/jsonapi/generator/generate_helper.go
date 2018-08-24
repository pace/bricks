// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package generator

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
)

func (g *Generator) addGoDoc(typeName, description string) {
	if description != "" {
		g.goSource.Comment(fmt.Sprintf("%s %s", typeName, description))
	} else {
		g.goSource.Comment(fmt.Sprintf("%s ...", typeName))
	}
}

func (g *Generator) goType(stmt *jen.Statement, schema *openapi3.Schema, tags map[string]string) error {
	switch schema.Type {
	case "string":
		switch schema.Format {
		case "byte": // TODO: needs to be base64 encoded/decoded
			stmt.Index().Byte()
		case "binary":
			stmt.Index().Byte()
		case "date-time":
			addValidator(tags, "rfc3339WithoutZone")
			stmt.Qual("time", "Time")
		case "date":
			addValidator(tags, "time(2006-01-02)")
			stmt.Qual("time", "Date")
		case "uuid":
			addValidator(tags, "uuid")
			stmt.String()
		default:
			stmt.String()
		}
	case "integer":
		switch schema.Format {
		case "int32":
			stmt.Int32()
		default:
			stmt.Int64()
		}
	case "number":
		switch schema.Format {
		case "float":
			stmt.Float32()
		case "double":
			fallthrough
		default:
			stmt.Float64()
		}
	case "boolean":
		stmt.Bool()
	case "array":
		g.goType(stmt.Index(), schema.Items.Value, tags)
	default:
		return fmt.Errorf("Unknown type: %s", schema.Type)
	}

	// add enum validation
	if len(schema.Enum) > 0 {
		strs := make([]string, len(schema.Enum))
		for i := 0; i < len(schema.Enum); i++ {
			strs[i] = fmt.Sprintf("%v", schema.Enum[i])
		}
		addValidator(tags, fmt.Sprintf("in(%v)", strings.Join(strs, ",")))
	}

	return nil
}

func (g *Generator) CommentOrExample(stmt *jen.Statement, schema *openapi3.Schema) {
	if schema.Description != "" {
		stmt.Comment(schema.Description)
	} else if schema.Example != nil {
		stmt.Comment(fmt.Sprintf("Example: \"%v\"", schema.Example))
	}
}

func addValidator(tags map[string]string, validator string) {
	cur := tags["valid"]
	if cur != "" {
		validator = tags["valid"] + "," + validator
	}
	tags["valid"] = validator
}
