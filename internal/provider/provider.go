package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	api "github.com/segmentio/public-api-sdk-go/api"
)

var _ provider.Provider = &SegmentProvider{}

type SegmentProvider struct {
	version string
}

type SegmentProviderModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

func (p *SegmentProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "segment"
	resp.Version = p.version
}

func (p *SegmentProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Segment Host",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Segment API Token",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *SegmentProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data SegmentProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Segment API Token",
			"The provider cannot create the Segment API client as there is an unknown configuration value for the Segment API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SEGMENT_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("SEGMENT_HOST")
	token := os.Getenv("SEGMENT_TOKEN")

	if !data.Host.IsNull() {
		host = data.Host.ValueString()
	}

	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddWarning(
			"Missing Segment Host",
			"The provider will default the host to api.segment.com. If your Segment workspace is in another region"+
				"Set the host value in the configuration or use the SEGMENT_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Segment API Token",
			"The provider cannot create the Segment API client as there is a missing or empty value for the Segment API token. "+
				"Set the token value in the configuration or use the SEGMENT_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	configuration := api.NewConfiguration()
	client := api.NewAPIClient(configuration)
	ctx = context.WithValue(context.Background(), api.ContextAccessToken, token)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *SegmentProvider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}

func (p *SegmentProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SegmentProvider{
			version: version,
		}
	}
}
