package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomuser"
	"github.com/folio-sec/terraform-provider-zoom/internal/provider/httpclient"
	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/folio-sec/terraform-provider-zoom/internal/provider/zoomclient"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/autoreceptionist"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/autoreceptionistivr"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/blockedlist"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/callhandling"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/callqueue"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/callqueuemember"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/callqueuephonenumber"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/callqueuepolicy"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/externalcontact"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/phonenumbers"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/sharedlinegroup"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/sharedlinegroupmember"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/sharedlinegroupphonenumber"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/site"
	phoneuser "github.com/folio-sec/terraform-provider-zoom/internal/services/phone/user"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/usercallingplans"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/phone/userphonenumber"
	"github.com/folio-sec/terraform-provider-zoom/internal/services/user/user"
	"github.com/folio-sec/terraform-provider-zoom/internal/zoomoauth"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

// Ensure zoomProvider satisfies various provider interfaces.
var _ provider.Provider = &ZoomProvider{}

type ZoomProvider struct {
	version      string
	ProviderData *shared.ProviderData
}

func (p *ZoomProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "zoom"
	resp.Version = p.version
}

func (p *ZoomProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
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

func (p *ZoomProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Zoom Phone API client")
	var config zoomProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	accountID := lo.TernaryF(config.AccountID.IsNull() || config.AccountID.IsUnknown(), func() string {
		return os.Getenv("ZOOM_ACCOUNT_ID")
	}, func() string {
		return config.AccountID.ValueString()
	})
	if accountID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("account_id"),
			"Missing Zoom Account ID",
			"The provider cannot create the Zoom API client as there is a missing or empty value for the Zoom Account ID. Please set the value in provider configuration or the ZOOM_ACCOUNT_ID environment variable. If either is already set, ensure the value is not empty.",
		)
	}

	clientID := lo.TernaryF(config.ClientID.IsNull() || config.ClientID.IsUnknown(), func() string {
		return os.Getenv("ZOOM_CLIENT_ID")
	}, func() string {
		return config.ClientID.ValueString()
	})
	if clientID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Missing Zoom Client ID",
			"The provider cannot create the Zoom API client as there is a missing or empty value for the Zoom Client ID. Please set the value in provider configuration or the ZOOM_CLIENT_ID environment variable. If either is already set, ensure the value is not empty.",
		)
	}

	clientSecret := lo.TernaryF(config.ClientSecret.IsNull() || config.ClientSecret.IsUnknown(), func() string {
		return os.Getenv("ZOOM_CLIENT_SECRET")
	}, func() string {
		return config.ClientSecret.ValueString()
	})
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

	tflog.Debug(ctx, "Creating Zoom Phone API client")

	retryClient := retryablehttp.NewClient().StandardClient()
	zoomOAuthClient, err := zoomoauth.NewClient(zoomoauth.WithHTTPClient(retryClient))
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

	httpClient := &http.Client{
		Transport: httpclient.NewNoJsonResponseRoundTripper(
			ctx,
			httpclient.NewLoggingRoundTripper(ctx, retryClient.Transport),
		),
	}

	zoomPhoneClient, err := zoomphone.NewClient(
		"https://api.zoom.us/v2",
		zoomclient.ZoomPhoneClientSecurity{
			AccessToken: res.AccessToken,
		},
		zoomphone.WithClient(httpClient),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Zoom Phone API client",
			fmt.Sprintf("An unexpected error occurred when creating the Zoom Phone API client. Error: %s", err.Error()),
		)
		return
	}

	zoomUserClient, err := zoomuser.NewClient(
		"https://api.zoom.us/v2",
		zoomclient.ZoomUserClientSecurity{
			AccessToken: res.AccessToken,
		},
		zoomuser.WithClient(httpClient),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Zoom User API client",
			fmt.Sprintf("An unexpected error occurred when creating the Zoom User API client. Error: %s", err.Error()),
		)
		return
	}

	p.ProviderData = &shared.ProviderData{
		PhoneClient: zoomPhoneClient,
		UserClient:  zoomUserClient,
	}

	resp.DataSourceData = p.ProviderData
	resp.ResourceData = p.ProviderData
}

func (p *ZoomProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		autoreceptionist.NewPhoneAutoReceptionistResource,
		autoreceptionistivr.NewPhoneAutoReceptionistIvrResource,
		blockedlist.NewPhoneBlockedListResource,
		callhandling.NewPhoneCallHandlingBusinessHoursResource,
		callhandling.NewPhoneCallHandlingClosedHoursResource,
		callhandling.NewPhoneCallHandlingHolidayHoursResource,
		callqueue.NewPhoneCallQueueResource,
		callqueuemember.NewPhoneCallQueueMembersResource,
		callqueuephonenumber.NewPhoneCallQueuePhoneNumbersResource,
		callqueuepolicy.NewPhoneCallQueuePolicyVoiceMailResource,
		externalcontact.NewPhoneExternalContactResource,
		sharedlinegroup.NewPhoneSharedLineGroupResource,
		sharedlinegroupmember.NewPhoneSharedLineGroupMembersResource,
		sharedlinegroupphonenumber.NewPhoneSharedLineGroupPhoneNumbersResource,
		phoneuser.NewPhoneUserResource,
		usercallingplans.NewPhoneUserCallingPlansResource,
		userphonenumber.NewPhoneUserPhoneNumbersResource,
		site.NewPhoneSiteResource,
	}
}

func (p *ZoomProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		autoreceptionist.NewPhoneAutoReceptionistDataSource,
		blockedlist.NewPhoneBlockedListDataSource,
		callqueue.NewPhoneCallQueueDataSource,
		phonenumbers.NewPhonePhoneNumbersDataSource,
		phoneuser.NewPhoneUsersDataSource,
		sharedlinegroup.NewPhoneSharedLineGroupDataSource,
		user.NewUsersDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ZoomProvider{
			version: version,
		}
	}
}
