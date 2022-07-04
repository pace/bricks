// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package generator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (g *Generator) addGoDoc(typeName, description string) {
	if description != "" {
		g.goSource.Comment(fmt.Sprintf("%s %s", typeName, description))
	} else {
		g.goSource.Comment(fmt.Sprintf("%s ...", typeName))
	}
}

func (g *Generator) goType(stmt *jen.Statement, schema *openapi3.Schema, tags map[string]string) *typeGenerator { // nolint: gocyclo
	return &typeGenerator{
		g:      g,
		stmt:   stmt,
		schema: schema,
		tags:   tags,
	}
}

type typeGenerator struct {
	g       *Generator
	stmt    *jen.Statement
	schema  *openapi3.Schema
	tags    map[string]string
	isParam bool
}

func (g *typeGenerator) invoke() error { // nolint: gocyclo
	switch g.schema.Type {
	case "string":
		switch g.schema.Format {
		case "byte": // TODO: needs to be base64 encoded/decoded
			g.stmt.Index().Byte()
		case "binary":
			g.stmt.Index().Byte()
		case "date-time":
			if jsonapi, ok := g.tags["jsonapi"]; ok { // add hint for jsonapi that time is in iso8601 format
				g.tags["jsonapi"] = jsonapi + ",iso8601"
			} else {
				addValidator(g.tags, "iso8601")
			}

			if g.isParam {
				g.stmt.Qual("time", "Time")
			} else {
				g.stmt.Op("*").Qual("time", "Time") // time.Time can not be nil, so a pointer is needed for omitempty to work
			}
		case "date":
			addValidator(g.tags, "time(2006-01-02)")
			if g.isParam {
				g.stmt.Qual("time", "Time")
			} else {
				g.stmt.Op("*").Qual("time", "Time") // time.Time can not be nil, so a pointer is needed for omitempty to work
			}
		case "uuid":
			addValidator(g.tags, "uuid")
			if g.schema.Nullable {
				g.stmt.Op("*").String()
			} else {
				g.stmt.String()
			}
		case "decimal":
			addValidator(g.tags, "matches(^(\\d*\\.)?\\d+$)")
			if g.isParam {
				g.stmt.Qual(pkgDecimal, "Decimal")
			} else {
				g.stmt.Op("*").Qual(pkgDecimal, "Decimal")
			}
		default:
			if g.schema.Nullable {
				g.stmt.Op("*").String()
			} else {
				g.stmt.String()
			}
		}
	case "integer":
		removeOmitempty(g.tags)
		switch g.schema.Format {
		case "int32":
			if g.schema.Nullable {
				g.stmt.Op("*").Int32()
			} else {
				g.stmt.Int32()
			}
		default:
			if g.schema.Nullable {
				g.stmt.Op("*").Int64()
			} else {
				g.stmt.Int64()
			}
		}
	case "number":
		switch g.schema.Format {
		case "decimal":
			if g.isParam {
				removeOmitempty(g.tags)
				g.stmt.Qual(pkgDecimal, "Decimal")
			} else {
				g.stmt.Op("*").Qual(pkgDecimal, "Decimal")
			}
		case "float":
			removeOmitempty(g.tags)
			if g.schema.Nullable {
				g.stmt.Op("*").Float32()
			} else {
				g.stmt.Float32()
			}
		case "double":
			fallthrough
		default:
			removeOmitempty(g.tags)
			if g.schema.Nullable {
				g.stmt.Op("*").Float64()
			} else {
				g.stmt.Float64()
			}
		}
	case "boolean":
		removeOmitempty(g.tags)
		if g.schema.Nullable {
			g.stmt.Op("*").Bool()
		} else {
			g.stmt.Bool()
		}
	case "array": // nolint: goconst
		removeOmitempty(g.tags)
		err := g.g.goType(g.stmt.Index(), g.schema.Items.Value, g.tags).invoke()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown type: %s", g.schema.Type)
	}

	// add enum validation
	if len(g.schema.Enum) > 0 {
		strs := make([]string, len(g.schema.Enum))
		for i := 0; i < len(g.schema.Enum); i++ {
			strs[i] = fmt.Sprintf("%v", g.schema.Enum[i])
		}

		// in case the field/value is optional
		// an empty value needs to be added to the enum validator
		if hasValidator(g.tags, "optional") {
			strs = append(strs, "")
		}

		addValidator(g.tags, fmt.Sprintf("in(%v)", strings.Join(strs, "|")))
	}

	return nil
}

func (g *Generator) commentOrExample(stmt *jen.Statement, schema *openapi3.Schema) {
	if schema.Description != "" {
		stmt.Comment(schema.Description)
	} else if schema.Example != nil {
		stmt.Comment(fmt.Sprintf("Example: \"%v\"", schema.Example))
	}
}

func hasSecuritySchema(swagger *openapi3.Swagger) bool {
	return len(swagger.Components.SecuritySchemes) > 0
}

func addValidator(tags map[string]string, validator string) {
	cur := tags["valid"]
	if cur != "" {
		validator = tags["valid"] + "," + validator
	}
	tags["valid"] = validator
}

func hasValidator(tags map[string]string, validator string) bool {
	validatorCfg, ok := tags["valid"]
	if !ok {
		return false
	}
	validators := strings.Split(validatorCfg, ",")
	for _, v := range validators {
		if strings.HasPrefix(v, validator) {
			return true
		}
	}

	return false
}

var idRegex = regexp.MustCompile("Id$")

func goNameHelper(name string) string {
	caser := cases.Title(language.Und, cases.NoLower)
	name = caser.String(name)
	name = strings.Replace(name, "Url", "URL", -1)
	name = idRegex.ReplaceAllString(name, "ID")
	return name
}

func nameFromSchemaRef(ref *openapi3.SchemaRef) string {
	name := goNameHelper(filepath.Base(ref.Ref))
	if name == "." {
		return ""
	}
	return name
}
