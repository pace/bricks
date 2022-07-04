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
	"github.com/ettle/strcase"
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
	pkgOIDC           = "github.com/pace/bricks/http/oidc"
	pkgApiKey         = "github.com/pace/bricks/http/security/apikey"
	pkgDecimal        = "github.com/shopspring/decimal"
)

const serviceInterface = "Service"
const rootRouterName = "router"
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
	sort.Stable(sort.StringSlice(keys))

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
		g.buildRouterHelpers,
		g.buildRouter,
		g.buildRouterWithFallbackAsArg,
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
	sort.Stable(sort.StringSlice(keys))

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
			tg := g.goType(paramStmt, param.Value.Schema.Value, tags)
			tg.isParam = true

			err := tg.invoke()
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
	for _, route := range routes {
		if err := g.buildSubServiceInterface(route, schema); err != nil {
			return err
		}
	}

	subServices := make([]jen.Code, 0)
	for _, route := range routes {
		subServices = append(subServices, jen.Id(generateSubServiceName(route.handler)))
	}

	g.goSource.Line()
	g.goSource.Line()
	g.goSource.Comment("Legacy Interface.")
	g.goSource.Comment("Use this if you want to fully implement a service.")
	g.goSource.Type().Id(serviceInterface).Interface(subServices...)

	return nil
}

func (g *Generator) buildSubServiceInterface(route *route, schema *openapi3.Swagger) error {
	methods := make([]jen.Code, 0)

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

	g.goSource.Line().Commentf("%s interface for %s handler", serviceInterface, route.handler)
	g.goSource.Type().Id(generateSubServiceName(route.handler)).Interface(methods...)

	return nil
}

func (g *Generator) buildRouter(routes []*route, schema *openapi3.Swagger) error {
	routerBody, err := g.buildRouterBodyWithFallback(routes, schema, jen.Id(rootRouterName).Dot("NotFoundHandler"))
	if err != nil {
		return nil
	}
	g.addGoDoc("Router", "implements: "+schema.Info.Title+"\n\n"+schema.Info.Description)
	serviceInterfaceVariable := jen.Id("service").Interface()
	if hasSecuritySchema(schema) {
		g.goSource.Func().Id("Router").Params(
			serviceInterfaceVariable, jen.Id("authBackend").Id(authBackendInterface)).Op("*").Qual(pkgGorillaMux, "Router").Block(routerBody...)

	} else {
		g.goSource.Func().Id("Router").Params(
			serviceInterfaceVariable).Op("*").Qual(pkgGorillaMux, "Router").Block(routerBody...)
	}
	return nil
}

func (g *Generator) buildRouterWithFallbackAsArg(routes []*route, schema *openapi3.Swagger) error {
	routerBody, err := g.buildRouterBodyWithFallback(routes, schema, jen.Id("fallback"))
	if err != nil {
		return nil
	}
	g.addGoDoc("Router", "implements: "+schema.Info.Title+"\n\n"+schema.Info.Description)
	serviceInterfaceVariable := jen.Id("service").Interface()
	if hasSecuritySchema(schema) {
		g.goSource.Func().Id("RouterWithFallback").Params(
			serviceInterfaceVariable, jen.Id("authBackend").Id(authBackendInterface), jen.Id("fallback").Qual("net/http", "Handler")).Op("*").Qual(pkgGorillaMux, "Router").Block(routerBody...)

	} else {
		g.goSource.Func().Id("RouterWithFallback").Params(
			serviceInterfaceVariable, jen.Id("fallback").Qual("net/http", "Handler")).Op("*").Qual(pkgGorillaMux, "Router").Block(routerBody...)
	}
	return nil
}

func (g *Generator) buildRouterHelpers(routes []*route, schema *openapi3.Swagger) error {
	needsSecurity := hasSecuritySchema(schema)

	// sort the routes with query parameter to the top
	sortableRoutes := sortableRouteList(routes)
	sort.Stable(&sortableRoutes)

	fallbackName := "fallback"
	fallback := jen.Id(fallbackName).Qual("net/http", "Handler")
	// add all route handlers
	for i := 0; i < len(sortableRoutes); i++ {
		route := sortableRoutes[i]
		var routeCallParams *jen.Statement
		if needsSecurity {
			routeCallParams = jen.List(jen.Id("service"), jen.Id("authBackend"))
		} else {
			routeCallParams = jen.List(jen.Id("service"))
		}
		primaryHandler := jen.Id(route.handler).Call(routeCallParams)
		fallbackHandler := jen.Id(fallbackName)
		ifElse := make([]jen.Code, 0)
		for _, handler := range []jen.Code{primaryHandler, fallbackHandler} {
			block := jen.Return(handler)
			ifElse = append(ifElse, block)
		}

		if len(ifElse) < 1 {
			panic("if-else slice should contain two elements, one with the service interface being called and one passing the NotFoundHandler")
		}

		implGuard := jen.If(
			jen.List(jen.Id("service"), jen.Id("ok")).Op(":=").Id("service").Assert(jen.Id(generateSubServiceName(route.handler))),
			jen.Id("ok")).Block(ifElse[0]).Else().Block(ifElse[1])

		comment := jen.Commentf("%s helper that checks if the given service fulfills the interface. Returns fallback handler if not, otherwise returns matching handler.", generateHandlerTypeAssertionHelperName(route.handler))

		var callParams *jen.Statement
		if needsSecurity {
			callParams = jen.List(jen.Id("service").Id("interface{}"), fallback, jen.Id("authBackend").Id(authBackendInterface))
		} else {
			callParams = jen.List(jen.Id("service").Id("interface{}"), fallback)
		}
		helper := jen.Func().Id(generateHandlerTypeAssertionHelperName(route.handler)).
			Params(callParams).Qual("net/http", "Handler").Block(implGuard).Line().Line()

		g.goSource.Line().Add(comment)
		g.goSource.Add(helper)
	}

	return nil
}

func (g *Generator) buildRouterBodyWithFallback(routes []*route, schema *openapi3.Swagger, fallback jen.Code) ([]jen.Code, error) {
	needsSecurity := hasSecuritySchema(schema)
	startInd := 0
	var routeStmts []jen.Code
	if needsSecurity {
		startInd++
		routeStmts = make([]jen.Code, 2, (len(routes)+2)*len(schema.Servers)+2)
		// Init Authentication
		var names []string
		for name := range schema.Components.SecuritySchemes {
			names = append(names, name)
		}
		sort.Stable(sort.StringSlice(names))
		for _, name := range names {
			routeStmts = append(routeStmts, jen.Id("authBackend").Dot("Init"+strings.Title(name)).Call(jen.Id("cfg"+strings.Title(name))))
		}
	} else {
		routeStmts = make([]jen.Code, 1, (len(routes)+2)*len(schema.Servers)+1)

	}
	// create new router
	routeStmts[startInd] = jen.Id(rootRouterName).Op(":=").Qual(pkgGorillaMux, "NewRouter").Call()

	// Note: we don't restrict host, scheme and port to ease development
	pathsIdx := make(map[string]struct{})
	var paths []string
	for _, server := range schema.Servers {
		serverUrl, err := url.Parse(server.URL)
		if err != nil {
			return nil, err
		}
		if _, ok := pathsIdx[serverUrl.Path]; !ok {
			paths = append(paths, serverUrl.Path)
		}
		pathsIdx[serverUrl.Path] = struct{}{}
	}

	// but generate subrouters for each server
	for i, path := range paths {
		subrouterID := fmt.Sprintf("s%d", i+1)

		// init and return the router
		routeStmts = append(routeStmts, jen.Comment(fmt.Sprintf("Subrouter %s - Path: %s", subrouterID, path)))
		routeStmts = append(routeStmts, jen.Id(subrouterID).Op(":=").Id("router").
			Dot("PathPrefix").Call(jen.Lit(path)).Dot("Subrouter").Call())

		// sort the routes with query parameter to the top
		sortableRoutes := sortableRouteList(routes)
		sort.Stable(&sortableRoutes)

		// add all route handlers
		for i := 0; i < len(sortableRoutes); i++ {
			route := sortableRoutes[i]
			var routeCallParams *jen.Statement
			if needsSecurity {
				routeCallParams = jen.List(jen.Id("service"), fallback, jen.Id("authBackend"))
			} else {
				routeCallParams = jen.List(jen.Id("service"), fallback)
			}
			helper := jen.Id(generateHandlerTypeAssertionHelperName(route.handler)).Call(routeCallParams)
			routeStmt := jen.Id(subrouterID).Dot("Methods").Call(jen.Lit(route.method)).
				Dot("Path").Call(jen.Lit(route.url.Path))

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

			routeStmt.Dot("Handler").Call(helper)

			routeStmts = append(routeStmts, routeStmt)

		}
	}

	// return
	routeStmts = append(routeStmts, jen.Return(jen.Id("router")))

	return routeStmts, nil
}

func (g *Generator) buildHandler(method string, op *openapi3.Operation, pattern string, pathItem *openapi3.PathItem, secSchemes map[string]*openapi3.SecuritySchemeRef) (*route, error) {
	needsSecurity := len(secSchemes) > 0
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
	// sanitise operationID
	oid = strcase.ToGoCamel(oid)
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
	if needsSecurity {
		if op.Security != nil {
			var err error
			auth, err = generateAuthorization(op, secSchemes)
			if err != nil {
				return nil, err
			}
		}
	}
	g.addGoDoc(handler, fmt.Sprintf("handles request/response marshaling and validation for \n %s %s",
		method, pattern))
	var params *jen.Statement
	if needsSecurity {
		params = jen.List(jen.Id("service").Id(generateSubServiceName(route.handler)), jen.Id("authBackend").Id(authBackendInterface))
	} else {
		params = jen.List(jen.Id("service").Id(generateSubServiceName(route.handler)))
	}
	g.goSource.Func().Id(handler).Params(params).Qual("net/http", "Handler").Block(
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
				).Line().Select().Block(
					jen.Case(jen.Op("<-").Id("ctx").Dot("Done").Call()),
					jen.If().Id("ctx").Dot("Err").Call().Op("!=").Nil().
						Block(
							jen.Comment("Context cancellation should not be reported if it's the request context"),
							jen.Id("w").Dot("WriteHeader").Call(jen.Lit(499)),
							jen.If().Id("err").Op("!=").Nil().Op("&&").Op("!").Parens(
								jen.Qual("errors", "Is").Call(jen.Id("err"), jen.Qual("context", "Canceled")).Op("||").
									Qual("errors", "Is").Call(jen.Id("err"), jen.Qual("context", "DeadlineExceeded")),
							).Block(
								jen.Comment("Report unclean error handling (err != context err) to sentry"),
								jen.Qual(pkgMaintErrors, "Handle").Call(jen.Id("ctx"), jen.Id("err")),
							),
						),
					jen.Default(),
					jen.If().Id("err").Op("!=").Nil().Block(
						jen.Qual(pkgMaintErrors, "HandleError").Call(jen.Id("err"),
							jen.Lit(handler),
							jen.Id("w"),
							jen.Id("r")),
					),
				)

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
	req := *op.Security
	r := &jen.Group{}
	if len(req[0]) == 0 {
		return r, nil
	}

	multipleSecSchemes := len(req[0]) > 1
	var err error
	if multipleSecSchemes {
		r, err = generateAuthorizationForMultipleSecSchemas(op, secSchemes)
	} else {
		r, err = generateAuthorizationForSingleSecSchema(op, secSchemes)
	}
	if err != nil {
		return nil, err
	}

	r.Line().Id("r").Op("=").Id("r.WithContext").Call(jen.Id("ctx"))
	return r, nil
}

func generateAuthorizationForSingleSecSchema(op *openapi3.Operation, schemas map[string]*openapi3.SecuritySchemeRef) (*jen.Group, error) {
	req := *op.Security
	r := &jen.Group{}
	if len(req[0]) == 0 {
		return nil, nil
	}
	for name, secConfig := range (*op.Security)[0] {
		securityScheme := schemas[name]
		switch securityScheme.Value.Type {
		case "oauth2", "openIdConnect":
			if len(secConfig) > 0 {
				r.Line().List(jen.Id("ctx"), jen.Id("ok")).Op(":=").Id("authBackend."+authFuncPrefix+strings.Title(name)).Call(jen.Id("r"), jen.Id("w"), jen.Lit(secConfig[0]))
			} else {
				r.Line().List(jen.Id("ctx"), jen.Id("ok")).Op(":=").Id("authBackend."+authFuncPrefix+strings.Title(name)).Call(jen.Id("r"), jen.Id("w"), jen.Lit(""))
			}
		case "apiKey":
			if len(secConfig) > 0 {
				return nil, fmt.Errorf("security config for api key authorization needs %d values but had: %d", 0, len(secConfig))
			}
			r.Line().List(jen.Id("ctx"), jen.Id("ok")).Op(":=").Id("authBackend."+authFuncPrefix+strings.Title(name)).Call(jen.Id("r"), jen.Id("w"))
		default:
			return nil, fmt.Errorf("security Scheme of type %q is not suppported", securityScheme.Value.Type)
		}
	}
	r.Line().If(jen.Op("!").Id("ok")).Block(jen.Return())
	return r, nil
}

func generateAuthorizationForMultipleSecSchemas(op *openapi3.Operation, secSchemes map[string]*openapi3.SecuritySchemeRef) (*jen.Group, error) {
	var orderedSec [][]string
	// Security Schemes are sorted for a reliable order of the code
	for name, val := range (*op.Security)[0] {
		vals := []string{name}
		orderedSec = append(orderedSec, append(vals, val...))
	}
	sort.Slice(orderedSec, func(i, j int) bool {
		return orderedSec[i][0] < orderedSec[j][0]
	})

	r := &jen.Group{}
	last := &jen.Group{}
	last.Qual("net/http", "Error").Call(jen.Id("w"), jen.Lit("Authorization Error"), jen.Qual("net/http", "StatusUnauthorized"))
	last.Line().Return()

	r.Line().Var().Id("ctx").Id("context.Context")
	r.Line().Var().Id("ok").Id("bool")
	for _, val := range orderedSec {
		name := val[0]
		securityScheme := secSchemes[name]
		innerBlock := &jen.Group{}
		innerBlock.Line().List(jen.Id("ctx"), jen.Id("ok")).Op("=").Id("authBackend." + authFuncPrefix + strings.Title(name))
		switch securityScheme.Value.Type {
		case "oauth2", "openIdConnect":
			if len(val) >= 2 {
				innerBlock.Call(jen.Id("r"), jen.Id("w"), jen.Lit(val[1]))
			} else {
				innerBlock.Call(jen.Id("r"), jen.Id("w"), jen.Lit(""))
			}
		case "apiKey":
			if len(val) > 1 {
				return nil, fmt.Errorf("security config for api key authorization needs %d values but had: %d", 0, len(val))
			}
			innerBlock.Call(jen.Id("r"), jen.Id("w"))
		default:
			return nil, fmt.Errorf("security Scheme of type %q is not suppported", securityScheme.Value.Type)
		}
		innerBlock.Line().If(jen.Op("!").Id("ok")).Block(jen.Return())
		r.Line().If(jen.Id("authBackend." + authCanAuthFuncPrefix + strings.Title(name))).Call(jen.Id("r")).Block(innerBlock).Else()
	}
	r.Block(last)
	return r, nil
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

func generateSubServiceName(handler string) string {
	return fmt.Sprintf("%s%s", handler, serviceInterface)
}

func generateHandlerTypeAssertionHelperName(handler string) string {
	return fmt.Sprintf("%sWithFallbackHelper", handler)
}
