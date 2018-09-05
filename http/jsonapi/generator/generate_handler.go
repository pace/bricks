// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package generator

import (
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	gorillaMux     = "github.com/gorilla/mux"
	httpJsonapi    = "lab.jamit.de/pace/go-microservice/http/jsonapi/runtime"
	jsonAPIMetrics = "lab.jamit.de/pace/go-microservice/maintenance/metrics/jsonapi"
	govalidator    = "github.com/asaskevich/govalidator"
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

	// TODO: maybe more 500 errors depending on the context result
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
		err := g.buildPath(pattern, path, &routes)
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

func (g *Generator) buildPath(pattern string, pathItem *openapi3.PathItem, routes *[]*route) error {
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

		route, err := g.buildHandler(op.method, op.operation, pattern, pathItem)
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
			return fmt.Errorf("Failed to parse response code %s: %v", code, err)
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
					jen.Qual(httpJsonapi, "WriteError").Call(
						jen.Id("w"),
						jen.Lit(codeNum),
						jen.Id("err"),
					),
				)
			}()
		} else if mt := response.Value.Content.Get(jsonapiContent); mt != nil {
			typeReference, err := g.generateTypeReference(route.serviceFunc+methodName,
				mt.Schema)
			if err != nil {
				return err
			}
			method.Params(typeReference)

			defer func() { // defer to put methods after type
				// generate the method as function for the implementing type
				g.addGoDoc(methodName, fmt.Sprintf("responds with jsonapi marshaled data (HTTP code %d)", codeNum))
				g.goSource.Func().Params(jen.Id("w").Op("*").Id(route.responseTypeImpl)).
					Id(methodName).Params(jen.Id("data").Add(typeReference)).Block(
					jen.Qual(httpJsonapi, "Marshal").Call(
						jen.Id("w"),
						jen.Id("data"),
						jen.Lit(codeNum),
					),
				)
			}()
		} else {
			method.Params()

			defer func() { // defer to put methods after type
				// generate the method as function for the implementing type
				g.addGoDoc(methodName, fmt.Sprintf("responds with empty response (HTTP code %d)", codeNum))
				g.goSource.Func().Params(jen.Id("w").Op("*").Id(route.responseTypeImpl)).
					Id(methodName).Params().Block(
					jen.Id("w").Dot("WriteHeader").Call(jen.Lit(codeNum)),
				)
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
			ref, err := g.generateTypeReference(route.serviceFunc+"Content", mt.Schema)
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
		g.addGoDoc(route.responseType, "is a standard http.Request extended with the\n"+
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

	g.goSource.Type().Id(serviceInterface).Interface(methods...)

	return nil
}

func (g *Generator) buildRouter(routes []*route, schema *openapi3.Swagger) error {
	routeStmts := make([]jen.Code, 1, (len(routes)+2)*len(schema.Servers)+1)

	// create new router
	routeStmts[0] = jen.Id("router").Op(":=").Qual(gorillaMux, "NewRouter").Call()

	// Note: we don't restrict host, scheme and port to ease development
	// but generate subrouters for each server
	for i, server := range schema.Servers {
		subrouterID := fmt.Sprintf("s%d", i+1)
		url, err := url.Parse(server.URL)
		if err != nil {
			return err
		}

		// init and return the router
		routeStmts = append(routeStmts, jen.Comment(fmt.Sprintf("Subrouter %s - %v", subrouterID, url)))
		routeStmts = append(routeStmts, jen.Id(subrouterID).Op(":=").Id("router").
			Dot("PathPrefix").Call(jen.Lit(url.Path)).Dot("Subrouter").Call())

		// sort the routes with query parameter to the top
		sortableRoutes := sortableRouteList(routes)
		sort.Sort(&sortableRoutes)

		// add all route handlers
		for i := 0; i < len(sortableRoutes); i++ {
			route := sortableRoutes[i]

			// generic route
			routeStmt := jen.Id(subrouterID).Dot("Methods").Call(jen.Lit(route.method)).
				Dot("Path").Call(jen.Lit(route.url.Path)).
				Dot("Handler").Call(jen.Id(route.handler).Call(jen.Id("service")))

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

	g.addGoDoc("Router", "implements: "+schema.Info.Title+"\n\n"+schema.Info.Description)
	g.goSource.Func().Id("Router").Params(
		jen.Id("service").Id(serviceInterface),
	).Op("*").Qual(gorillaMux, "Router").Block(routeStmts...)

	return nil
}

func (g *Generator) buildHandler(method string, op *openapi3.Operation, pattern string, pathItem *openapi3.PathItem) (*route, error) {
	route := &route{
		method:    strings.ToUpper(method),
		pattern:   pattern,
		operation: op,
	}

	// avoid ruby style path parameters
	if strings.Contains(pattern, "/:") {
		log.Printf("Note: Don't use ruby style path parameters: %s", pattern)
	}

	// use OperationID for go function names or generate the name
	oid := strings.Title(op.OperationID)
	if oid == "" {
		log.Printf("Note: Avoid automatic method name generation for path (use OperationID): %s", pattern)
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
	g.addGoDoc(handler, fmt.Sprintf("handles request/response marshaling and validation for \n %s %s",
		method, pattern))
	g.goSource.Func().Id(handler).Params(
		jen.Id("service").Id("Service"),
	).Qual("net/http", "Handler").Block(
		jen.Return().Qual("net/http", "HandlerFunc").Call(
			jen.Func().Params(
				jen.Id("w").Qual("net/http", "ResponseWriter"),
				jen.Id("r").Op("*").Qual("net/http", "Request"),
			).BlockFunc(func(g *jen.Group) {
				// recover panics
				// TODO: add more context and send to sentry, return error code
				// that can be correlated with the client
				g.Defer().Func().Call().BlockFunc(func(g *jen.Group) {
					g.If(jen.Id("r").Op(":=").Id("recover").Call().Op(";").Id("r").Op("!=").Nil()).Block(
						jen.Qual("fmt", "Printf").Call(jen.Lit("Panic %s: %v\n"), jen.Lit(handler), jen.Id("r")),
						jen.Qual("runtime/debug", "PrintStack").Call(),
						jen.Qual(httpJsonapi, "WriteError").Call(
							jen.Id("w"),
							jen.Qual("net/http", "StatusInternalServerError"),
							// don't leak info about the internal panic
							jen.Qual("errors", "New").Call(jen.Lit("Error")),
						),
					)
				}).Call()

				// response writer
				g.Id("writer").Op(":=").Id(route.responseTypeImpl).
					Block(jen.Id("ResponseWriter").Op(":").
						Qual(jsonAPIMetrics, "NewMetric").Call(
						jen.Lit(gen.serviceName),
						jen.Lit(route.pattern),
						jen.Id("w"),
						jen.Id("r")).Op(","))

				// request
				g.Id("request").Op(":=").Id(route.requestType).
					Block(jen.Id("Request").Op(":").Id("r").Op(","))

				// vars in case parameters are given
				if len(route.operation.Parameters) > 0 {
					g.Id("vars").Op(":=").Qual(gorillaMux, "Vars").Call(jen.Id("r"))

					// all parameters need to be parsed
					g.If().Op("!").Qual(httpJsonapi, "ScanParameters").CallFunc(func(g *jen.Group) {
						g.Id("w")
						g.Id("r")

						for _, param := range route.operation.Parameters {
							name := generateParamName(param)
							g.Op("&").Qual(httpJsonapi, "ScanParameter").Block(
								jen.Id("Data").Op(":").Op("&").Id("request").Dot(name).Op(","),
								jen.Id("Location").Op(":").Qual(httpJsonapi, "ScanIn"+strings.Title(param.Value.In)).Op(","),
								jen.Id("Input").Op(":").Id("vars").Index(jen.Lit(param.Value.Name)).Op(","),
								jen.Id("Name").Op(":").Lit(param.Value.Name).Op(","),
							)
						}
					}).Block(jen.Return())
				}

				// validate parameters / body
				if requestBody || len(route.operation.Parameters) > 0 {
					g.If().Op("!").Qual(httpJsonapi, "ValidateParameters").Call(
						jen.Id("w"),
						jen.Id("r"),
						jen.Op("&").Id("request"),
					).Block(
						jen.Return().Comment("invalid request stop further processing"),
					)
				}

				// invoke service and handle error with internal server error response
				invokeService := jen.Id("err").Op(":=").Id("service").Dot(route.serviceFunc).Call(
					jen.Id("r").Dot("Context").Call(),
					jen.Op("&").Id("writer"),
					jen.Op("&").Id("request"),
				).Line().If().Id("err").Op("!=").Nil().Block(
					// TODO: add more context and send to sentry
					jen.Qual(httpJsonapi, "WriteError").Call(
						jen.Id("w"),
						jen.Qual("net/http", "StatusInternalServerError"),
						jen.Id("err"),
					),
				)

				// if there is a request body unmarshal it then call the service
				// otherwise directly call the service
				if requestBody {
					g.If(jen.Qual(httpJsonapi, "Unmarshal").Call(
						jen.Id("w"),
						jen.Id("r"),
						jen.Op("&").Id("request").Dot("Content"))).Block(invokeService)
				} else {
					g.Add(invokeService)
				}
			}),
		),
	)

	return route, nil
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
	return "Param" + goNameHelper(param.Value.Name)
}
