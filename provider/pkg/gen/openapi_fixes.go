package gen

import (
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

// FixOpenAPIDoc applies patches to the raw OpenAPI spec
// before passing it to pulschema.
func FixOpenAPIDoc(openAPIDoc *openapi3.T) error {
	fixMachineMetadataKeyEndpoint(openAPIDoc)
	fixMachineUpdateEndpoint(openAPIDoc)
	fixVolumeEndpoint(openAPIDoc)

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

func fixMachineUpdateEndpoint(openAPIDoc *openapi3.T) {
	pathItem := openAPIDoc.Paths.Find("/apps/{app_name}/machines/{machine_id}")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/machines/{machine_id}")

	contract.Assertf(pathItem.Post != nil, "Expected POST operation on /apps/{app_name}/machines/{machine_id}")
	pathItem.Post.OperationID = "Machine_update"
}

func fixVolumeEndpoint(openAPIDoc *openapi3.T) {
	pathItem := openAPIDoc.Paths.Find("/apps/{app_name}/volumes/{volume_id}")
	contract.Assertf(pathItem != nil, "Expected to find request path /apps/{app_name}/volumes/{volume_id}")

	contract.Assertf(pathItem.Get != nil, "Expected GET operation on /apps/{app_name}/volumes/{volume_id}")
	pathItem.Get.OperationID = "Volume_get"

	contract.Assertf(pathItem.Put != nil, "Expected PUT operation on /apps/{app_name}/volumes/{volume_id}")
	pathItem.Put.OperationID = "Volume_update"
}
