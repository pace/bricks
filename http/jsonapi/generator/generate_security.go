// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/11/05 by Charlotte Pröller

package generator

import (
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pace/bricks/maintenance/errors"
)

const (
	authBackendInterface = "AuthorizationBackend"
	authFuncPrefix       = "Authorize"
)

// buildSecurityBackendInterface builds the interface that is used to do the authentication.
// It creates one method for each security type and an init method for handling the securityConfigs.
// The Methods are named AuthenticateNAME and Init.
func (g *Generator) buildSecurityBackendInterface(schema *openapi3.Swagger) error {
	securitySchemes := schema.Components.SecuritySchemes
	// r contains the methods for the security interface
	r := &jen.Group{}
	// configs contains the names and types of the needed configs for the init method
	// (that initializes the backend with the security configs)
	var configs []jen.Code

	// Because the order of the values while iterating over a map is randomized the generated result can only be tested if the keys are sorted
	var keys []string
	for k := range securitySchemes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		value := securitySchemes[name]
		r.Line().Id(authFuncPrefix + strings.Title(name))
		switch value.Value.Type {
		case "oauth2":
			configs = append(configs, jen.Id("cfg"+strings.Title(name)).Op("*").Qual(pkgOAuth2, "Config"))
			r.Params(jen.Id("r").Id("*http.Request"), jen.Id("w").Id("http.ResponseWriter"), jen.Id("scope").String())
		case "apiKey":
			configs = append(configs, jen.Id("cfg"+strings.Title(name)).Op("*").Qual(pkgApiKey, "Config"))
			r.Params(jen.Id("r").Id("*http.Request"), jen.Id("w").Id("http.ResponseWriter"))
		default:
			return errors.New("security schema type not supported: " + value.Value.Type)
		}
		r.Params(jen.Id("context.Context"), jen.Id("bool"))
	}

	r.Line().Id("Init").Params(configs...)
	g.goSource.Type().Id(authBackendInterface).Interface(r)
	return nil
}

// BuildSecurityConfigs creates structs with the config of each security schema
func (g *Generator) buildSecurityConfigs(schema *openapi3.Swagger) error {
	securitySchemes := schema.Components.SecuritySchemes
	// Because the order of the values while iterating over a map is randomized the generated result can only be tested if the keys are sorted
	var keys []string
	for k := range securitySchemes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		value := securitySchemes[name]
		instanceVal := jen.Dict{}
		var pkgName string
		switch value.Value.Type {
		case "oauth2":
			pkgName = pkgOAuth2
			t := value.Value.Description
			instanceVal[jen.Id("Description")] = jen.Lit(t)
			// If any flow is present create the flow struct
			if value.Value.Flows != nil {
				flows := map[string]*openapi3.OAuthFlow{
					"AuthorizationCode": value.Value.Flows.AuthorizationCode,
					"ClientCredentials": value.Value.Flows.ClientCredentials,
					"Implicit":          value.Value.Flows.Implicit,
					"Password":          value.Value.Flows.Password}
				for flowname, flow := range flows {
					if flow != nil {
						instanceVal[jen.Id(flowname)] = jen.Op("&").Qual(pkgOAuth2, "Flow").Values(getValuesFromFlow(flow))
					}
				}
			}

		case "apiKey":
			pkgName = pkgApiKey
			instanceVal[jen.Id("Description")] = jen.Lit(value.Value.Description)
			instanceVal[jen.Id("In")] = jen.Lit(value.Value.In)
			instanceVal[jen.Id("Name")] = jen.Lit(value.Value.Name)
		default:
			return errors.New("security schema type not supported: " + value.Value.Type)
		}
		g.goSource.Var().Id("cfg"+strings.Title(name)).Op("=").Op("&").Qual(pkgName, "Config").Values(instanceVal)
	}
	return nil
}

// getValuesFromFlow puts the values from the OAuth Flow in a jen.Dict to generate it
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
