// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package generator

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pace/bricks/maintenance/log"
)

const (
	pkgGorillaMux     = "github.com/gorilla/mux"
	pkgJSONAPIRuntime = "github.com/pace/bricks/http/jsonapi/runtime"
	pkgJSONAPIMetrics = "github.com/pace/bricks/maintenance/metric/jsonapi"
	pkgMaintErrors    = "github.com/pace/bricks/maintenance/errors"
	pkgOpentracing    = "github.com/opentracing/opentracing-go"
	pkgOAuth2         = "github.com/pace/bricks/http/oauth2"
	pkgApiKey         = "github.com/pace/bricks/http/security/apikey"
)

const serviceInterface = "Service"
const jsonapiContent = "application/vnd.api+json"

var noValidation = map[string]string{"valid": "-"}

// List of responses that will be handled on the framework level and
// are therefore not handled by the user
var generatorResponseBlacklist = map[string]bool{
	"401": true, // if no bearer token is provided
	"406": true, // if accept header is unacceptable
	"415": true, // if media type is invalid
	"422": true, // handled by go-validator
	"500": true, // if service returns an error

	// TODO(vil): maybe more 500 errors depending on the context result
	// e.g. Temporary errors (implementing the temporary interface)
	// result in retry later / also rate limiting
}

type routeGeneratorFunc func([]*route, *openapi3.Swagger) error

// BuildHandler generates the request handlers based on gorilla mux
func (g *Generator) BuildHandler(schema *openapi3.Swagger) error {
	paths := schema.Paths
	// sort by key
	keys := make([]string, 0, len(paths))
	for k := range paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var routes []*route

	for _, pattern := range keys {
		path := paths[pattern]
		err := g.buildPath(pattern, path, &routes, schema.Components.SecuritySchemes)
		if err != nil {
			return err
		}
	}

	funcs := []routeGeneratorFunc{
		g.generateRequestResponseTypes,
		g.buildServiceInterface,
		g.buildRouter,
	}
	for _, fn := range funcs {
		err := fn(routes, schema)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) buildPath(pattern string, pathItem *openapi3.PathItem, routes *[]*route, secSchemes map[string]*openapi3.SecuritySchemeRef) error {
	operations := []struct {
		method    string
		operation *openapi3.Operation
	}{
		{"Connect", pathItem.Connect},
		{"Delete", pathItem.Delete},
		{"Get", pathItem.Get},
		{"Head", pathItem.Head},
		{"Options", pathItem.Options},
		{"Patch", pathItem.Patch},
		{"Post", pathItem.Post},
		{"Put", pathItem.Put},
		{"Trace", pathItem.Trace},
	}

	for _, op := range operations {
		// since the list contains all operations some can be nil
		if op.operation == nil {
			continue
		}

		route, err := g.buildHandler(op.method, op.operation, pattern, pathItem, secSchemes)
		if err != nil {
			return err
		}

		err = route.parseURL()
		if err != nil {
			return err
		}

		*routes = append(*routes, route)
	}

	return nil
}

func (g *Generator) generateRequestResponseTypes(routes []*route, schema *openapi3.Swagger) error {
	for _, route := range routes {
		// generate ...ResponseWriter for each route
		err := g.generateResponseInterface(route, schema)
		if err != nil {
			return err
		}

		// generate ...Request for each route
		err = g.generateRequestStruct(route, schema)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateResponseInterface(route *route, schema *openapi3.Swagger) error {
	var methods []jen.Code
	methods = append(methods, jen.Qual("net/http", "ResponseWriter"))

	// sort by key
	keys := make([]string, 0, len(route.operation.Responses))
	for k := range route.operation.Responses {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, code := range keys {
		response := route.operation.Responses[code]

		// don't generate response helpers for things that are handled by the framework
		if generatorResponseBlacklist[code] {
			continue
		}

		// error responses have an error message parameter
		codeNum, err := strconv.Atoi(code)
		if err != nil {
			return fmt.Errorf("failed to parse response code %s: %v", code, err)
		}

		// generate method name
		var methodName string
		if response.Ref != "" {
			methodName = generateMethodName(filepath.Base(response.Ref))
		} else {
			methodName = generateMethodName(response.Value.Description)
		}

		method := jen.Id(methodName)
		if codeNum >= 400 {
			method.Params(jen.Error())

			defer func() { // defer to put methods after type
				// generate the method as function for the implementing type
				g.addGoDoc(methodName, fmt.Sprintf("responds with jsonapi error (HTTP code %d)", codeNum))
				g.goSource.Func().Params(jen.Id("w").Op("*").Id(route.responseTypeImpl)).
					Id(methodName).Params(jen.Id("err").Error()).Block(
					jen.Qual(pkgJSONAPIRuntime, "WriteError").Call(
						jen.Id("w"),
						jen.Lit(codeNum),
						jen.Id("err"),
					),
				)
			}()
		} else if mt := response.Value.Content.Get(jsonapiContent); mt != nil {
			typeReference, err := g.generateTypeReference(route.serviceFunc+methodName,
				mt.Schema, false)
			if err != nil {
				return err
			}
			method.Params(typeReference)

			defer func() { // defer to put methods after type
				// generate the method as function for the implementing type
				g.addGoDoc(methodName, fmt.Sprintf("responds with jsonapi marshaled data (HTTP code %d)", codeNum))
				g.goSource.Func().Params(jen.Id("w").Op("*").Id(route.responseTypeImpl)).
					Id(methodName).Params(jen.Id("data").Add(typeReference)).Block(
					jen.Qual(pkgJSONAPIRuntime, "Marshal").Call(
						jen.Id("w"),
						jen.Id("data"),
						jen.Lit(codeNum),
					),
				)
			}()
		} else {
			method.Params()

			defer func() { // defer to put methods after type
				// get mime type if any
				mime := "application/vnd.api+json"
				for m := range response.Value.Content {
					mime = m
					break // only the first mime type will be respected
				}

				// generate the method as function for the implementing type
				g.addGoDoc(methodName, fmt.Sprintf("responds with empty response (HTTP code %d)", codeNum))
				g.goSource.Func().Params(jen.Id("w").Op("*").Id(route.responseTypeImpl)).
					Id(methodName).Params().BlockFunc(func(g *jen.Group) {
					// set the content type for the response (prevents the go guess work -> improves performance)
					g.Id("w").Dot("Header").Call().Dot("Set").Call(jen.Lit("Content-Type"), jen.Lit(mime))
					g.Id("w").Dot("WriteHeader").Call(jen.Lit(codeNum))
				})
			}()
		}

		methods = append(methods, method)
	}

	// Comment and type
	g.addGoDoc(route.responseType, "is a standard http.ResponseWriter extended with methods\n"+
		"to generate the respective responses easily")
	g.goSource.Type().Id(route.responseType).Interface(methods...)

	// Implementation type
	g.goSource.Type().Id(route.responseTypeImpl).Struct(
		jen.Qual("net/http", "ResponseWriter"),
	)

	return nil
}

func (g *Generator) generateRequestStruct(route *route, schema *openapi3.Swagger) error {
	body := route.operation.RequestBody
	var fields []jen.Code

	// add http request
	fields = append(fields, jen.Id("Request").Op("*").Qual("net/http", "Request").Tag(noValidation))

	// add request type
	if body != nil {
		if mt := body.Value.Content.Get(jsonapiContent); mt != nil {
			ref, err := g.generateTypeReference(route.serviceFunc+"Content", mt.Schema, true)
			if err != nil {
				return err
			}
			// the content has noValidation to only check parametes first
			// then unmarshal and then check the content after
			fields = append(fields, jen.Id("Content").Add(ref).Tag(noValidation))
		}
	}

	// add parameters
	for _, param := range route.operation.Parameters {
		paramName := generateParamName(param)
		paramStmt := jen.Id(paramName)
		tags := make(map[string]string)
		if param.Value.Required {
			tags["valid"] = "required"
		} else {
			tags["valid"] = "optional"
		}

		// add go type
		if param.Value.Schema.Ref != "" {
			paramStmt.Id(goNameHelper(filepath.Base(param.Value.Schema.Ref)))
		} else {
			err := g.goType(paramStmt, param.Value.Schema.Value, tags)
			if err != nil {
				return err
			}
		}

		fields = append(fields, paramStmt.Tag(tags))
	}

	// add comment and generate type
	if body != nil {
		g.addGoDoc(route.requestType, body.Value.Description)
	} else {
		g.addGoDoc(route.requestType, "is a standard http.Request extended with the\n"+
			"un-marshaled content object")
	}
	g.goSource.Type().Id(route.requestType).Struct(fields...)

	return nil
}

func (g *Generator) buildServiceInterface(routes []*route, schema *openapi3.Swagger) error {
	methods := make([]jen.Code, len(routes))

	for _, route := range routes {
		if route.operation.Description != "" {
			methods = append(methods, jen.Comment(fmt.Sprintf("%s %s\n\n%s", route.serviceFunc, route.operation.Summary, route.operation.Description)))
		} else {
			methods = append(methods, jen.Comment(fmt.Sprintf("%s %s", route.serviceFunc, route.operation.Summary)))
		}
		methods = append(methods, jen.Id(route.serviceFunc).Params(
			jen.Qual("context", "Context"),
			jen.Id(route.responseType),
			jen.Op("*").Id(route.requestType),
		).Id("error"))
	}

	g.goSource.Line().Commentf("%s interface for all handlers", serviceInterface)
	g.goSource.Type().Id(serviceInterface).Interface(methods...)

	return nil
}

func (g *Generator) buildRouter(routes []*route, schema *openapi3.Swagger) error {
	routeStmts := make([]jen.Code, 2, (len(routes)+2)*len(schema.Servers)+2)
	// Init Authentication
	var configs []jen.Code
	for name := range schema.Components.SecuritySchemes {
		configs = append(configs, jen.Id("cfg"+strings.Title(name)))
	}
	routeStmts[0] = jen.If(jen.Id("authBackend").Op("!=").Nil()).Block(jen.Id("authBackend").Dot("Init").Call(configs...))
	// create new router
	routeStmts[1] = jen.Id("router").Op(":=").Qual(pkgGorillaMux, "NewRouter").Call()

	// Note: we don't restrict host, scheme and port to ease development
	paths := make(map[string]struct{})
	for _, server := range schema.Servers {
		serverUrl, err := url.Parse(server.URL)
		if err != nil {
			return err
		}
		paths[serverUrl.Path] = struct{}{}
	}

	// but generate subrouters for each server
	i := 0
	for path := range paths {
		subrouterID := fmt.Sprintf("s%d", i+1)

		// init and return the router
		routeStmts = append(routeStmts, jen.Comment(fmt.Sprintf("Subrouter %s - Path: %s", subrouterID, path)))
		routeStmts = append(routeStmts, jen.Id(subrouterID).Op(":=").Id("router").
			Dot("PathPrefix").Call(jen.Lit(path)).Dot("Subrouter").Call())

		// sort the routes with query parameter to the top
		sortableRoutes := sortableRouteList(routes)
		sort.Sort(&sortableRoutes)

		// add all route handlers
		for i := 0; i < len(sortableRoutes); i++ {
			route := sortableRoutes[i]

			// generic route
			routeStmt := jen.Id(subrouterID).Dot("Methods").Call(jen.Lit(route.method)).
				Dot("Path").Call(jen.Lit(route.url.Path)).
				Dot("Handler").Call(jen.Id(route.handler).Call(jen.List(jen.Id("service"), jen.Id("authBackend"))))

			// add query parameters for route matching
			if len(route.queryValues) > 0 {
				for key, value := range route.queryValues {
					if len(value) != 1 {
						panic("query paths can only handle one query parameter with the same name!")
					}
					routeStmt.Dot("Queries").Call(jen.Lit(key), jen.Lit(value[0]))
				}
			}

			// add the name to build routes
			routeStmt.Dot("Name").Call(jen.Lit(route.serviceFunc))

			routeStmts = append(routeStmts, routeStmt)
		}
	}

	// return
	routeStmts = append(routeStmts, jen.Return(jen.Id("router")))

	g.addGoDoc("RouterWithAuthentication", "implements: "+schema.Info.Title+"\n\n"+schema.Info.Description)
	serviceInterfaceVariable := jen.Id("service").Id(serviceInterface)
	g.goSource.Func().Id("RouterWithAuthentication").Params(
		serviceInterfaceVariable, jen.Id("authBackend").Id(authBackendInterface)).Op("*").Qual(pkgGorillaMux, "Router").Block(routeStmts...)

	block := jen.Return(jen.Id("RouterWithAuthentication").Call(jen.Id("service"), jen.Nil()))

	g.addGoDocDeprecated("Router", "kept for backward compatibility. Please use RouteWithAuthentication, Remove the Middleware and implement the AuthenticationBackend")

	g.goSource.Func().Id("Router").Params(
		serviceInterfaceVariable).Op("*").Qual(pkgGorillaMux, "Router").Block(block)

	return nil
}

func (g *Generator) buildHandler(method string, op *openapi3.Operation, pattern string, pathItem *openapi3.PathItem, secSchemes map[string]*openapi3.SecuritySchemeRef) (*route, error) {
	route := &route{
		method:    strings.ToUpper(method),
		pattern:   pattern,
		operation: op,
	}

	// avoid ruby style path parameters
	if strings.Contains(pattern, "/:") {
		log.Warnf("Note: Don't use ruby style path parameters: %s", pattern)
	}

	// use OperationID for go function names or generate the name
	oid := strings.Title(op.OperationID)
	if oid == "" {
		log.Warnf("Note: Avoid automatic method name generation for path (use OperationID): %s", pattern)
		oid = generateName(method, op, pattern)
	}
	handler := oid + "Handler"
	route.handler = handler
	route.serviceFunc = oid
	route.responseType = oid + "ResponseWriter"
	route.responseTypeImpl = strings.ToLower(oid[:1]) + oid[1:] + "ResponseWriter"
	route.requestType = oid + "Request"

	// check if handler has request body
	var requestBody bool
	if body := op.RequestBody; body != nil {
		if mt := body.Value.Content.Get(jsonapiContent); mt != nil {
			requestBody = true
		}
	}

	// generate handler function
	gen := g // generator is used less frequent then the jen group, make available with longer name
	var auth *jen.Group
	if op.Security != nil {
		var err error
		auth, err = generateAuthorization(op, secSchemes)
		if err != nil {
			return nil, err
		}
	}
	g.addGoDoc(handler, fmt.Sprintf("handles request/response marshaling and validation for \n %s %s",
		method, pattern))
	g.goSource.Func().Id(handler).Params(
		jen.Id("service").Id(serviceInterface), jen.Id("authBackend").Id(authBackendInterface)).Qual("net/http", "Handler").Block(
		jen.Return().Qual("net/http", "HandlerFunc").Call(
			jen.Func().Params(
				jen.Id("w").Qual("net/http", "ResponseWriter"),
				jen.Id("r").Op("*").Qual("net/http", "Request"),
			).BlockFunc(func(g *jen.Group) {
				// recover panics
				g.Defer().Qual(pkgMaintErrors, "HandleRequest").Call(jen.Lit(handler), jen.Id("w"), jen.Id("r"))

				g.Add(auth)

				// set tracing context
				g.Line().Comment("Trace the service function handler execution")
				g.List(jen.Id("handlerSpan"), jen.Id("ctx")).Op(":=").Qual(pkgOpentracing, "StartSpanFromContext").Call(
					jen.Id("r").Dot("Context").Call(), jen.Lit(handler))
				g.Defer().Id("handlerSpan").Dot("Finish").Call()
				g.Line().Comment("Setup context, response writer and request type")

				// response writer
				g.Id("writer").Op(":=").Id(route.responseTypeImpl).
					Block(jen.Id("ResponseWriter").Op(":").
						Qual(pkgJSONAPIMetrics, "NewMetric").Call(
						jen.Lit(gen.serviceName),
						jen.Lit(route.pattern),
						jen.Id("w"),
						jen.Id("r")).Op(","))

				// request
				g.Id("request").Op(":=").Id(route.requestType).
					Block(jen.Id("Request").Op(":").Id("r").Dot("WithContext").Call(jen.Id("ctx")).Op(","))

				// vars in case parameters are given
				g.Line().Comment("Scan and validate incoming request parameters")
				if len(route.operation.Parameters) > 0 {
					// path parameters need the vars
					needVars := false
					for _, param := range route.operation.Parameters {
						if param.Value.In == "path" {
							needVars = true
						}
					}
					if needVars {
						g.Id("vars").Op(":=").Qual(pkgGorillaMux, "Vars").Call(jen.Id("r"))
					}

					// all parameters need to be parsed
					g.If().Op("!").Qual(pkgJSONAPIRuntime, "ScanParameters").CallFunc(func(g *jen.Group) {
						g.Id("w")
						g.Id("r")

						for _, param := range route.operation.Parameters {
							name := generateParamName(param)
							g.Op("&").Qual(pkgJSONAPIRuntime, "ScanParameter").BlockFunc(func(g *jen.Group) {
								g.Id("Data").Op(":").Op("&").Id("request").Dot(name).Op(",")
								g.Id("Location").Op(":").Qual(pkgJSONAPIRuntime, "ScanIn"+strings.Title(param.Value.In)).Op(",")
								if param.Value.In == "path" {
									g.Id("Input").Op(":").Id("vars").Index(jen.Lit(param.Value.Name)).Op(",")
								}
								g.Id("Name").Op(":").Lit(param.Value.Name).Op(",")
							})
						}
					}).Block(jen.Return())
				}

				// validate parameters / body
				if requestBody || len(route.operation.Parameters) > 0 {
					g.If().Op("!").Qual(pkgJSONAPIRuntime, "ValidateParameters").Call(
						jen.Id("w"),
						jen.Id("r"),
						jen.Op("&").Id("request"),
					).Block(
						jen.Return().Comment("invalid request stop further processing"),
					)
				}

				// invoke service and handle error with internal server error response
				invokeService := jen.Comment("Invoke service that implements the business logic").Line().
					Id("err").Op(":=").Id("service").Dot(route.serviceFunc).Call(
					jen.Id("ctx"),
					jen.Op("&").Id("writer"),
					jen.Op("&").Id("request"),
				).Line().If().Id("err").Op("!=").Nil().Block(
					jen.Qual(pkgMaintErrors, "HandleError").Call(jen.Id("err"),
						jen.Lit(handler),
						jen.Id("w"),
						jen.Id("r")))

				// if there is a request body unmarshal it then call the service
				// otherwise directly call the service
				if requestBody {
					g.Line().Comment("Unmarshal the service request body")
					isArray := false
					mt := op.RequestBody.Value.Content.Get(jsonapiContent)
					if mt != nil {
						data := mt.Schema.Value.Properties["data"]
						if data != nil && data.Value.Type == "array" {
							if data.Ref != "" && data.Value.Items.Ref != "" {
								isArray = true
							}
						}
					}
					if isArray {
						typeName := nameFromSchemaRef(mt.Schema.Value.Properties["data"].Value.Items)
						g.List(jen.Id("ok"), jen.Id("data")).Op(":=").
							Qual(pkgJSONAPIRuntime, "UnmarshalMany").
							Call(
								jen.Id("w"),
								jen.Id("r"),
								jen.Qual("reflect", "TypeOf").Call(jen.New(jen.Id(typeName))),
							)
						g.If(jen.Id("ok")).Block(
							jen.Comment("Move the data"),
							jen.For(jen.List(jen.Id("_"), jen.Id("elem")).Op(":=").Range().Call(jen.Id("data"))).
								Block(
									jen.Id("request").Dot("Content").
										Op("=").
										Append(
											jen.Id("request").Dot("Content"),
											jen.Id("elem").Assert(jen.Id("*"+typeName)),
										),
								),
							invokeService,
						)
					} else {
						g.If(jen.Qual(pkgJSONAPIRuntime, "Unmarshal").Call(
							jen.Id("w"),
							jen.Id("r"),
							jen.Op("&").Id("request").Dot("Content"))).Block(invokeService)
					}
				} else {
					g.Line().Add(invokeService)
				}
			}),
		),
	)

	return route, nil
}

func generateAuthorization(op *openapi3.Operation, secSchemes map[string]*openapi3.SecuritySchemeRef) (*jen.Group, error) {
	r := &jen.Group{}
	r.Add(jen.Comment("Authentication Handling "))
	for _, sl := range *op.Security {
		for name, val := range sl {
			securityScheme := secSchemes[name]
			switch securityScheme.Value.Type {
			case "oauth2":
				r.Line().Add(jen.Comment("OAuth2 Authentication"))
				if len(val) < 1 {
					return nil, fmt.Errorf("security config for OAuth2 authorization needs %d values but had: %d", 1, len(val))
				}
				scope := val[0]
				r.Line().List(jen.Id("ctx"), jen.Id("ok")).Op(":=").Id("authBackend."+authFuncPrefix+strings.Title(name)).Call(jen.Id("r"), jen.Id("w"), jen.Lit(scope))
				r.Line().If(jen.Op("!").Id("ok")).Block(jen.Comment("No Error Handling needed, this is already done"), jen.Return())
				r.Line().Id("r").Op("=").Id("r.WithContext").Call(jen.Id("ctx"))
			case "apiKey":
				if len(val) > 0 {
					return nil, fmt.Errorf("security config for api key authoritzation needs %d values but had: %d", 0, len(val))
				}
				r.Line().List(jen.Id("ctx"), jen.Id("ok")).Op(":=").Id("authBackend."+authFuncPrefix+strings.Title(name)).Call(jen.Id("r"), jen.Id("w"))
				r.Line().If(jen.Op("!").Id("ok")).Block(jen.Comment("No Error Handling needed, this is already done"), jen.Return())
				r.Line().Id("r").Op("=").Id("r.WithContext").Call(jen.Id("ctx"))
			default:
				return nil, fmt.Errorf("security Scheme of type %q is not suppported", securityScheme.Value.Type)
			}
		}
	}
	result := &jen.Group{}
	result.Comment("Only do Authentication if a authentication Backend is available. Otherwise this is handled somewhere else")
	result.Line().If(jen.Id("authBackend").Op("!=").Nil()).Block(r)
	return result, nil
}

var asciiName = regexp.MustCompile("([^a-zA-Z]+)")

func generateName(method string, op *openapi3.Operation, pattern string) string {
	name := method
	parts := strings.Split(asciiName.ReplaceAllString(pattern, "/"), "/")
	for _, part := range parts {
		name += goNameHelper(part)
	}

	return goNameHelper(name)
}

func generateMethodName(description string) string {
	parts := strings.Split(asciiName.ReplaceAllString(description, " "), " ")
	for i := 0; i < len(parts); i++ {
		parts[i] = goNameHelper(parts[i])
	}
	return goNameHelper(strings.Join(parts, ""))
}

func generateParamName(param *openapi3.ParameterRef) string {
	return "Param" + generateMethodName(param.Value.Name)
}
