// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/11/05 by Charlotte Pröller

package generator

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pace/bricks/maintenance/errors"
)

const (
	authBackendInterface  = "AuthorizationBackend"
	authFuncPrefix        = "Authorize"
	authCanAuthFuncPrefix = "CanAuthorize"
)

// buildSecurityBackendInterface builds the interface that is used to do the authentication.
// It creates one method for each security type and an init method for handling the securityConfigs.
// The Methods are named AuthenticateNAME and Init.
func (g *Generator) buildSecurityBackendInterface(schema *openapi3.Swagger) error {
	if !hasSecuritySchema(schema) {
		return nil
	}
	securitySchemes := schema.Components.SecuritySchemes
	// r contains the methods for the security interface
	r := &jen.Group{}

	// Because the order of the values while iterating over a map is randomized the generated result can only be tested if the keys are sorted
	var keys []string
	for k := range securitySchemes {
		keys = append(keys, k)
	}
	sort.Stable(sort.StringSlice(keys))
	hasDuplicatedSecuritySchema := false
	for _, pathItem := range schema.Paths {
		for _, op := range pathItem.Operations() {
			if op.Security != nil {
				hasDuplicatedSecuritySchema = hasDuplicatedSecuritySchema || len((*op.Security)[0]) > 1
			}
		}
	}

	for _, name := range keys {
		value := securitySchemes[name]
		r.Line().Id(authFuncPrefix + strings.Title(name))
		switch value.Value.Type {
		case "oauth2":
			r.Params(jen.Id("r").Id("*http.Request"), jen.Id("w").Id("http.ResponseWriter"), jen.Id("scope").String()).Params(jen.Id("context.Context"), jen.Id("bool"))
			r.Line().Id("Init" + strings.Title(name)).Params(jen.Id("cfg"+strings.Title(name)).Op("*").Qual(pkgOAuth2, "Config"))
		case "openIdConnect":
			r.Params(jen.Id("r").Id("*http.Request"), jen.Id("w").Id("http.ResponseWriter"), jen.Id("scope").String()).Params(jen.Id("context.Context"), jen.Id("bool"))
			r.Line().Id("Init" + strings.Title(name)).Params(jen.Id("cfg"+strings.Title(name)).Op("*").Qual(pkgOIDC, "Config"))
		case "apiKey":
			r.Params(jen.Id("r").Id("*http.Request"), jen.Id("w").Id("http.ResponseWriter")).Params(jen.Id("context.Context"), jen.Id("bool"))
			r.Line().Id("Init" + strings.Title(name)).Params(jen.Id("cfg"+strings.Title(name)).Op("*").Qual(pkgApiKey, "Config"))
		case "http":
			r.Params(jen.Id("r").Id("*http.Request"), jen.Id("w").Id("http.ResponseWriter"), jen.Id("scope").String()).Params(jen.Id("context.Context"), jen.Id("bool"))
			r.Line().Id("Init" + strings.Title(name)).Params(jen.Id("cfg"+strings.Title(name)).Op("*").Qual(pkgHttp, "Config"))
		default:
			return errors.New("security schema type not supported: " + value.Value.Type)
		}

		if hasDuplicatedSecuritySchema {
			r.Line().Id(authCanAuthFuncPrefix + strings.Title(name)).Params(jen.Id("r").Id("*http.Request")).Id("bool")
		}
	}

	g.goSource.Type().Id(authBackendInterface).Interface(r)
	return nil
}

// BuildSecurityConfigs creates structs with the config of each security schema
func (g *Generator) buildSecurityConfigs(schema *openapi3.Swagger) error {
	if !hasSecuritySchema(schema) {
		return nil
	}
	securitySchemes := schema.Components.SecuritySchemes
	// Because the order of the values while iterating over a map is randomized the generated result can only be tested if the keys are sorted
	var keys []string
	for k := range securitySchemes {
		keys = append(keys, k)
	}
	sort.Stable(sort.StringSlice(keys))

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
		case "openIdConnect":
			pkgName = pkgOIDC
			instanceVal[jen.Id("Description")] = jen.Lit(value.Value.Description)
			if e, ok := value.Value.Extensions["openIdConnectUrl"]; ok {
				var url string
				if data, ok := e.(json.RawMessage); ok {
					err := json.Unmarshal(data, &url)
					if err != nil {
						return err
					}
					instanceVal[jen.Id("OpenIdConnectURL")] = jen.Lit(url)
				}
			}
		case "apiKey":
			pkgName = pkgApiKey
			instanceVal[jen.Id("Description")] = jen.Lit(value.Value.Description)
			instanceVal[jen.Id("In")] = jen.Lit(value.Value.In)
			instanceVal[jen.Id("Name")] = jen.Lit(value.Value.Name)
		case "http":
			pkgName = pkgHttp
			instanceVal[jen.Id("Description")] = jen.Lit(value.Value.Description)
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
