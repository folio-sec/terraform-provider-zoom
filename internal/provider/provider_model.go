package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type zoomProviderModel struct {
	AccountID    types.String `tfsdk:"account_id"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}
