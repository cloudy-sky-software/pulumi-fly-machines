package gen

import (
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

// FixOpenAPIDoc applies patches to the raw OpenAPI spec
// before passing it to pulschema.
func FixOpenAPIDoc(openAPIDoc *openapi3.T) error {
	fixMachineMetadataKeyEndpoint(openAPIDoc)

	return nil
}

func fixMachineMetadataKeyEndpoint(openAPIDoc *openapi3.T) {
	pathItem := openAPIDoc.Paths.Find("/apps/{app_name}/machines/{machine_id}/metadata/{key}")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/machines/{machine_id}/metadata/{key}")

	contract.Assertf(pathItem.Post != nil, "Expected POST operation on /apps/{app_name}/machines/{machine_id}/metadata/{key}")
	pathItem.Post.OperationID = "Machines_create_metadata_key"

	contract.Assertf(pathItem.Delete != nil, "Expected DELETE operation on /apps/{app_name}/machines/{machine_id}/metadata/{key}")
	pathItem.Delete.OperationID = "Machines_delete_metadata_key"
}
