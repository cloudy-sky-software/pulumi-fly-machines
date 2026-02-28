package gen

import (
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

func newPathParam(name, description string) *openapi3.ParameterRef {
	return &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        name,
			In:          "path",
			Description: description,
			Required:    true,
			Schema:      openapi3.NewStringSchema().NewRef(),
		},
	}
}

func appNameParam() *openapi3.ParameterRef {
	return newPathParam("app_name", "Fly App Name")
}

func hostnameParam() *openapi3.ParameterRef {
	return newPathParam("hostname", "Host name of the Fly App")
}

func ipParam() *openapi3.ParameterRef {
	return newPathParam("ip", "IP address")
}

// addMissingPathParams adds missing path parameters to operations
// on the given path item. The parameters are added to all operations
// that exist on the path item.
func addMissingPathParams(pathItem *openapi3.PathItem, params ...*openapi3.ParameterRef) {
	operations := map[string]*openapi3.Operation{
		"GET":    pathItem.Get,
		"PUT":    pathItem.Put,
		"POST":   pathItem.Post,
		"DELETE": pathItem.Delete,
		"PATCH":  pathItem.Patch,
	}

	for _, op := range operations {
		if op == nil {
			continue
		}
		op.Parameters = append(op.Parameters, params...)
	}
}

func fixCertificateEndpoints(openAPIDoc *openapi3.T) {
	// /apps/{app_name}/certificates - GET is missing app_name path param.
	pathItem := openAPIDoc.Paths.Find("/apps/{app_name}/certificates")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/certificates")
	addMissingPathParams(pathItem, appNameParam())

	// /apps/{app_name}/certificates/acme - POST is missing app_name path param.
	pathItem = openAPIDoc.Paths.Find("/apps/{app_name}/certificates/acme")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/certificates/acme")
	addMissingPathParams(pathItem, appNameParam())

	// /apps/{app_name}/certificates/custom - POST is missing app_name path param.
	pathItem = openAPIDoc.Paths.Find("/apps/{app_name}/certificates/custom")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/certificates/custom")
	addMissingPathParams(pathItem, appNameParam())

	// /apps/{app_name}/certificates/{hostname} - GET and DELETE are missing app_name and hostname path params.
	pathItem = openAPIDoc.Paths.Find("/apps/{app_name}/certificates/{hostname}")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/certificates/{hostname}")
	addMissingPathParams(pathItem, appNameParam(), hostnameParam())

	// /apps/{app_name}/certificates/{hostname}/acme - DELETE is missing app_name and hostname path params.
	pathItem = openAPIDoc.Paths.Find("/apps/{app_name}/certificates/{hostname}/acme")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/certificates/{hostname}/acme")
	addMissingPathParams(pathItem, appNameParam(), hostnameParam())

	// /apps/{app_name}/certificates/{hostname}/check - PUT is missing app_name and hostname path params.
	pathItem = openAPIDoc.Paths.Find("/apps/{app_name}/certificates/{hostname}/check")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/certificates/{hostname}/check")
	addMissingPathParams(pathItem, appNameParam(), hostnameParam())

	// /apps/{app_name}/certificates/{hostname}/custom - DELETE is missing app_name and hostname path params.
	pathItem = openAPIDoc.Paths.Find("/apps/{app_name}/certificates/{hostname}/custom")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/certificates/{hostname}/custom")
	addMissingPathParams(pathItem, appNameParam(), hostnameParam())
}

func fixIPAssignmentEndpoints(openAPIDoc *openapi3.T) {
	// /apps/{app_name}/ip_assignments - GET and POST are missing app_name path param.
	pathItem := openAPIDoc.Paths.Find("/apps/{app_name}/ip_assignments")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/ip_assignments")
	addMissingPathParams(pathItem, appNameParam())

	// /apps/{app_name}/ip_assignments/{ip} - DELETE is missing app_name and ip path params.
	pathItem = openAPIDoc.Paths.Find("/apps/{app_name}/ip_assignments/{ip}")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/ip_assignments/{ip}")
	addMissingPathParams(pathItem, appNameParam(), ipParam())
}

func fixMachineMetadataEndpoint(openAPIDoc *openapi3.T) {
	// PATCH /apps/{app_name}/machines/{machine_id}/metadata is missing a request body.
	// It should accept a map of strings.
	pathItem := openAPIDoc.Paths.Find("/apps/{app_name}/machines/{machine_id}/metadata")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/machines/{machine_id}/metadata")
	contract.Assertf(pathItem.Patch != nil, "Expected PATCH operation on /apps/{app_name}/machines/{machine_id}/metadata")

	pathItem.Patch.RequestBody = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().
			WithJSONSchema(openapi3.NewObjectSchema().
				WithAdditionalProperties(openapi3.NewStringSchema())),
	}

	// POST /apps/{app_name}/machines/{machine_id}/metadata/{key} is missing a request body.
	// It should accept an object with a single "value" property of type string.
	keyPathItem := openAPIDoc.Paths.Find("/apps/{app_name}/machines/{machine_id}/metadata/{key}")
	contract.Assertf(keyPathItem != nil, "Expected to find request path /apps/{app_name}/machines/{machine_id}/metadata/{key}")
	contract.Assertf(keyPathItem.Post != nil, "Expected POST operation on /apps/{app_name}/machines/{machine_id}/metadata/{key}")

	keyPathItem.Post.RequestBody = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().
			WithJSONSchema(openapi3.NewObjectSchema().
				WithProperty("value", openapi3.NewStringSchema())),
	}
}

// FixOpenAPIDoc applies patches to the raw OpenAPI spec
// before passing it to pulschema.
func FixOpenAPIDoc(openAPIDoc *openapi3.T) error {
	fixCertificateEndpoints(openAPIDoc)
	fixIPAssignmentEndpoints(openAPIDoc)
	fixMachineMetadataEndpoint(openAPIDoc)

	return nil
}
