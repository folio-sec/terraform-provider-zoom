package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	autoreceptionist2 "github.com/folio-sec/terraform-provider-zoom/internal/services/phone/autoreceptionist"
	"github.com/folio-sec/terraform-provider-zoom/internal/zoomoauth"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure zoomProvider satisfies various provider interfaces.
var _ provider.Provider = &zoomProvider{}

type zoomProvider struct {
	version string
}

type zoomProviderModel struct {
	AccountID    types.String `tfsdk:"account_id"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

type clientSecurity struct {
	AccessToken string
}

func (c clientSecurity) OpenapiAuthorization(_ context.Context, _ string) (zoomphone.OpenapiAuthorization, error) {
	return zoomphone.OpenapiAuthorization{}, nil
}
func (c clientSecurity) OpenapiOAuth(_ context.Context, _ string) (zoomphone.OpenapiOAuth, error) {
	return zoomphone.OpenapiOAuth{Token: c.AccessToken}, nil
}

func (p *zoomProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "zoom"
	resp.Version = p.version
}

func (p *zoomProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `The Zoom provider is used to interact with the resources.
The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources and data sources.

## Authentication

The Zoom provider offers a flexible means of providing credentials for authentication. The following methods are supported, in this order, and explained below:

- Environment variables
- Provider Config
`,
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				MarkdownDescription: "The Account ID for Zoom. This can also be sourced from the ZOOM_ACCOUNT_ID environment variable.",
				Optional:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The Client ID for Zoom. This can also be sourced from the ZOOM_CLIENT_ID environment variable.",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The Client Secret for Zoom. This can also be sourced from the ZOOM_CLIENT_SECRET environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *zoomProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Zoom Phone API client")
	var config zoomProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	accountID := os.Getenv("ZOOM_ACCOUNT_ID")
	clientID := os.Getenv("ZOOM_CLIENT_ID")
	clientSecret := os.Getenv("ZOOM_CLIENT_ID")

	if !config.AccountID.IsNull() || !config.AccountID.IsUnknown() {
		accountID = config.AccountID.ValueString()
	}
	if accountID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("account_id"),
			"Missing Zoom Account ID",
			"The provider cannot create the Zoom API client as there is a missing or empty value for the Zoom Account ID. Please set the value in provider configuration or the ZOOM_ACCOUNT_ID environment variable. If either is already set, ensure the value is not empty.",
		)
	}

	if !config.ClientID.IsNull() || !config.ClientID.IsUnknown() {
		clientID = config.ClientID.ValueString()
	}
	if clientID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Missing Zoom Client ID",
			"The provider cannot create the Zoom API client as there is a missing or empty value for the Zoom Client ID. Please set the value in provider configuration or the ZOOM_CLIENT_ID environment variable. If either is already set, ensure the value is not empty.",
		)
	}

	if !config.ClientSecret.IsNull() {
		clientSecret = config.ClientSecret.ValueString()
	}
	if clientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Missing Zoom Client Secret",
			"The provider cannot create the Zoom API client as there is a missing or empty value for the Zoom Client Secret. Please set the value in provider configuration or the ZOOM_CLIENT_SECRET environment variable. If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "account_id", accountID)
	ctx = tflog.SetField(ctx, "client_id", clientID)
	ctx = tflog.SetField(ctx, "client_secret", clientSecret)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "client_secret")

	tflog.Debug(ctx, "Creating Zoom Phone Master API client")

	retryableClient := retryablehttp.NewClient()
	httpClient := retryableClient.StandardClient()

	zoomOAuthClient, err := zoomoauth.NewClient(zoomoauth.WithHTTPClient(httpClient))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Zoom OAuth client",
			fmt.Sprintf("An unexpected error occurred when creating the Zoom OAuth client. Error: %s", err.Error()),
		)
		return
	}

	res, err := zoomOAuthClient.GetAccessToken(ctx, accountID, clientID, clientSecret)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get Zoom API access token",
			fmt.Sprintf("Unabled to get access token. Please also check your Account ID, Client ID or Client Secret just to be sure. Error: %s", err.Error()),
		)
		return
	}

	zoomPhoneMasterClient, err := zoomphone.NewClient(
		"https://api.zoom.us/v2",
		clientSecurity{
			AccessToken: res.AccessToken,
		},
		zoomphone.WithClient(httpClient),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Zoom Phone Master API client",
			fmt.Sprintf("An unexpected error occurred when creating the Zoom Phone Master API client. Error: %s", err.Error()),
		)
		return
	}

	providerData := &shared.ProviderData{
		PhoneMasterClient: zoomPhoneMasterClient,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *zoomProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		autoreceptionist2.NewPhoneReceptionistResource,
	}
}

func (p *zoomProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		autoreceptionist2.NewPhoneAutoReceptionistDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &zoomProvider{
			version: version,
		}
	}
}
