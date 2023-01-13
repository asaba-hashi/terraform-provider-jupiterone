package jupiterone

import (
	// "errors"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/jupiterone/terraform-provider-jupiterone/jupiterone/internal/client"
)

// JupiterOneProvider contains the initialized API client to communicate with the JupiterOne API
type JupiterOneProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
	Client  *client.JupiterOneClient
}

type JupiterOneProviderModel struct {
	APIKey    basetypes.StringValue `tfsdk:"api_key"`
	AccountID basetypes.StringValue `tfsdk:"account_id"`
	Region    basetypes.StringValue `tfsdk:"region"`

	// httpClient should _only_ be used for tests using go-vcr/cassettes
	httpClient *http.Client
}

var _ provider.Provider = &JupiterOneProvider{}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &JupiterOneProvider{
			version: version,
		}
	}
}

// Configure implements provider.Provider
func (p *JupiterOneProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data JupiterOneProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if p.Client == nil {
		// Check environment variables. Performing this as part of Configure is
		// the current de-facto way of "merging" defaults:
		// https://github.com/hashicorp/terraform-plugin-framework/issues/539#issuecomment-1334470425
		var err error
		p.Client, err = data.Client(ctx)

		if err != nil {
			resp.Diagnostics.AddError("failed to create JupiterOne client in provider configuration: %s", err.Error())
			return
		}
		log.Println("[INFO] JupiterOne client successfully initialized")
	} else {
		log.Println("[INFO] Using already configured client")
	}

	resp.DataSourceData = p
	resp.ResourceData = p
}

// NewClient configures the J1 client itself from the provider model and
// allows overrides from the env variables.
//
// TODO: For testing, will also accept an `http_client` set in the context for
// for use with `cassettes` tests. This is a result of updating to the new
// terraform-plugin-framework and needing to inject the client, while not
// duplicating code or messing up the existing cassettes.
func (m *JupiterOneProviderModel) Client(ctx context.Context) (*client.JupiterOneClient, error) {
	apiKey := os.Getenv("JUPITERONE_API_KEY")
	accountId := os.Getenv("JUPITERONE_ACCOUNT_ID")
	region := os.Getenv("JUPITERONE_REGION")

	if apiKey == "" {
		apiKey = m.APIKey.ValueString()
	}
	if accountId == "" {
		accountId = m.AccountID.ValueString()
	}
	if region == "" {
		region = m.Region.ValueString()
	}

	config := client.JupiterOneClientConfig{
		APIKey:     apiKey,
		AccountID:  accountId,
		Region:     region,
		HTTPClient: m.httpClient,
	}

	return config.Client()
}

// DataSources implements provider.Provider
func (*JupiterOneProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// Metadata implements provider.Provider
func (p *JupiterOneProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "jupiterone"
	resp.Version = p.version
}

// Resources implements provider.Provider
func (*JupiterOneProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewQuestionResource,
		NewQuestionRuleResource,
	}
}

// Schema implements provider.Provider
func (*JupiterOneProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				// TODO: needs to be optional to use env vars in Configure
				Optional:    true,
				Description: "API Key used to make requests to the JupiterOne APIs",
				Sensitive:   true,
			},
			"account_id": schema.StringAttribute{
				// TODO: needs to be optional to use env vars in Configure
				Optional:    true,
				Description: "JupiterOne account ID to create resources in",
			},
			"region": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}
