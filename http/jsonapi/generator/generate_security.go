// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/11/05 by Charlotte Pröller

package generator

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
)

// BuildSecurityConfigs builds structs for the securitySchemes. Builds variables of the security Schemes
// with the given values from the scheme. These Config variables are used for the authorization in each
// method.
func (g *Generator) BuildSecurityConfigs(schema *openapi3.Swagger) error {
	securitySchemes := schema.Components.SecuritySchemes
	_ = securitySchemes
	flowCreated := false
	for name, value := range securitySchemes {
		instanceVal := jen.Dict{}
		typeVal := &jen.Group{}

		t := value.Value.Type
		instanceVal[jen.Id("Type")] = jen.Lit(t)
		typeVal.Line().Id("Type").String().Tag(map[string]string{"json": "type,omitempty"})
		t = value.Value.BearerFormat
		addValue(t, "bearerFormat", typeVal, instanceVal)
		t = value.Value.Description
		addValue(t, "description", typeVal, instanceVal)
		t = value.Value.In
		addValue(t, "in", typeVal, instanceVal)
		t = value.Value.Scheme
		addValue(t, "scheme", typeVal, instanceVal)

		// If any flow is present create the flow struct
		if value.Value.Flows != nil {
			if !flowCreated {
				g.addStructForFlow()
				flowCreated = true
			}
			// create the variable for each flow that is present.
			aC := value.Value.Flows.AuthorizationCode
			cC := value.Value.Flows.ClientCredentials
			imp := value.Value.Flows.Implicit
			pass := value.Value.Flows.Password
			flows := []*openapi3.OAuthFlow{aC, cC, imp, pass}
			for _, flow := range flows {
				if flow != nil {
					typeVal.Line().Id("FlowAuthorizationCode").Id("*OAuth2Flow")
					instanceVal[jen.Id("FlowAuthorizationCode")] = jen.Op("&").Id("OAuth2Flow").Values(getValuesFromFlow(flow))
				}
			}
		}
		//add struct for the authentication scheme
		schemaStructName := strings.Title(name) + "Config"
		g.goSource.Type().Id(schemaStructName).Struct(typeVal)
		//add Variable with given values
		g.goSource.Var().Id("cfg" + name).Op("=").Op("&").Id(schemaStructName).Values(instanceVal)
	}
	return nil
}

//
func getValuesFromFlow(flow *openapi3.OAuthFlow) jen.Dict {
	r := jen.Dict{}
	r[jen.Id("AuthorizationURL")] = jen.Lit(flow.AuthorizationURL)
	r[jen.Id("TokenURL")] = jen.Lit(flow.TokenURL)
	r[jen.Id("RefreshURL")] = jen.Lit(flow.RefreshURL)
	scopes := jen.Dict{}
	for scope, descr := range flow.Scopes {
		scopes[jen.Lit(scope)] = jen.Lit(descr)
	}
	r[jen.Id("Scopes")] = jen.Map(jen.String()).String().Values(scopes)
	return r
}

// addValue if a value from the scheme is present the variable is added to the jen.Group of the struct and the
// value to the dictionary of values for the config variable of the security type
func addValue(val string, name string, structGroup *jen.Group, values jen.Dict) {
	if val != "" {
		//Use Title to make the Variable accessible
		varName := strings.Title(name)
		structGroup.Line().Id(varName).String().Tag(map[string]string{"json": name + ",omitempty"})
		values[jen.Id(varName)] = jen.Lit(val)
	}
}

// addStructForFlows creates a struct for flow information.
// Only supports the standard values, not any extensions,  only created once and only if needed
func (g *Generator) addStructForFlow() {
	val := &jen.Group{}
	val.Line().Id("AuthorizationURL").String().Tag(map[string]string{"json": "authorizationUrl,omitempty"})
	val.Line().Id("TokenURL").String().Tag(map[string]string{"json": "tokenUrl,omitempty"})
	val.Line().Id("RefreshURL").String().Tag(map[string]string{"json": "refreshUrl,omitempty"})
	val.Line().Id("Scopes").Map(jen.String()).String().Tag(map[string]string{"json": "scopes"})
	g.goSource.Type().Id("OAuth2Flow").Struct(val)
}
