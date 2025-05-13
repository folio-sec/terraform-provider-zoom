package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

type TestData struct {
	// ResourceName is the resource label (local name), e.g. "test".
	ResourceLabel string

	// ResourceName is the qualified resource name, e.g. "zoom_phone_user.test".
	ResourceName string

	// TerraformResourceType is the Terraform resource type, e.g. "zoom_phone_user".
	TerraformResourceType string
}

func NewTestData(_ *testing.T, tfResourceType, resourceLabel string) TestData {
	return TestData{
		ResourceLabel:         resourceLabel,
		ResourceName:          fmt.Sprintf("%[1]s.%[2]s", tfResourceType, resourceLabel),
		TerraformResourceType: tfResourceType,
	}
}

func (td TestData) RandomStringOfLength(len int) string {
	return acctest.RandString(len)
}
