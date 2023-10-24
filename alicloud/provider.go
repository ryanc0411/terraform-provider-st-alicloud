package alicloud

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	alicloudAdbClient "github.com/alibabacloud-go/adb-20190315/v2/client"
	alicloudDnsClient "github.com/alibabacloud-go/alidns-20150109/v4/client"
	alicloudBaseClient "github.com/alibabacloud-go/bssopenapi-20171214/v3/client"
	alicloudCdnClient "github.com/alibabacloud-go/cdn-20180510/v2/client"
	alicloudCmsClient "github.com/alibabacloud-go/cms-20190101/v8/client"
	alicloudCsClient "github.com/alibabacloud-go/cs-20151215/v4/client"
	alicloudOpenapiClient "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	alicloudAntiddosClient "github.com/alibabacloud-go/ddoscoo-20200101/v2/client"
	alicloudEmrClient "github.com/alibabacloud-go/emr-20210320/client"
	alicloudRamClient "github.com/alibabacloud-go/ram-20150501/v2/client"
	alicloudSlbClient "github.com/alibabacloud-go/slb-20140515/v4/client"

	"github.com/alibabacloud-go/tea/tea"
)

// Wrapper of AliCloud client
type alicloudClients struct {
	baseClient     *alicloudBaseClient.Client
	cdnClient      *alicloudCdnClient.Client
	antiddosClient *alicloudAntiddosClient.Client
	slbClient      *alicloudSlbClient.Client
	dnsClient      *alicloudDnsClient.Client
	ramClient      *alicloudRamClient.Client
	cmsClient      *alicloudCmsClient.Client
	adbClient      *alicloudAdbClient.Client
	emrClient      *alicloudEmrClient.Client
	csClient       *alicloudCsClient.Client
}

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &alicloudProvider{}
)

// New is a helper function to simplify provider server
func New() provider.Provider {
	return &alicloudProvider{}
}

type alicloudProvider struct{}

type alicloudProviderModel struct {
	Region    types.String `tfsdk:"region"`
	AccessKey types.String `tfsdk:"access_key"`
	SecretKey types.String `tfsdk:"secret_key"`
}

// Metadata returns the provider type name.
func (p *alicloudProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "st-alicloud"
}

// Schema defines the provider-level schema for configuration data.
func (p *alicloudProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Alibaba Cloud provider is used to interact with the many resources supported by Alibaba Cloud. " +
			"The provider needs to be configured with the proper credentials before it can be used.",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "Region for AliCloud API. May also be provided via ALICLOUD_REGION environment variable.",
				Optional:    true,
			},
			"access_key": schema.StringAttribute{
				Description: "Access Key for AliCloud API. May also be provided via ALICLOUD_ACCESS_KEY environment variable",
				Optional:    true,
			},
			"secret_key": schema.StringAttribute{
				Description: "Secret key for AliCloud API. May also be provided via ALICLOUD_SECRET_KEY environment variable",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a AliCloud API client for data sources and resources.
func (p *alicloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config alicloudProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.Region.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("region"),
			"Unknown AliCloud region",
			"The provider cannot create the AliCloud API client as there is an unknown configuration value for the"+
				"AliCloud API region. Set the value statically in the configuration, or use the ALICLOUD_REGION environment variable.",
		)
	}

	if config.AccessKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_key"),
			"Unknown AliCloud access key",
			"The provider cannot create the AliCloud API client as there is an unknown configuration value for the"+
				"AliCloud API access key. Set the value statically in the configuration, or use the ALICLOUD_ACCESS_KEY environment variable.",
		)
	}

	if config.SecretKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret_key"),
			"Unknown AliCloud secret key",
			"The provider cannot create the AliCloud API client as there is an unknown configuration value for the"+
				"AliCloud secret key. Set the value statically in the configuration, or use the ALICLOUD_SECRET_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	var region, accessKey, secretKey string
	if !config.Region.IsNull() {
		region = config.Region.ValueString()
	} else {
		region = os.Getenv("ALICLOUD_REGION")
	}

	if !config.AccessKey.IsNull() {
		accessKey = config.AccessKey.ValueString()
	} else {
		accessKey = os.Getenv("ALICLOUD_ACCESS_KEY")
	}

	if !config.SecretKey.IsNull() {
		secretKey = config.SecretKey.ValueString()
	} else {
		secretKey = os.Getenv("ALICLOUD_SECRET_KEY")
	}

	// If any of the expected configuration are missing, return
	// errors with provider-specific guidance.
	if region == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("region"),
			"Missing AliCloud API region",
			"The provider cannot create the AliCloud API client as there is a "+
				"missing or empty value for the AliCloud API region. Set the "+
				"region value in the configuration or use the ALICLOUD_REGION "+
				"environment variable. If either is already set, ensure the value "+
				"is not empty.",
		)
	}

	if accessKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_key"),
			"Missing AliCloud API access key",
			"The provider cannot create the AliCloud API client as there is a "+
				"missing or empty value for the AliCloud API access key. Set the "+
				"access key value in the configuration or use the ALICLOUD_ACCESS_KEY "+
				"environment variable. If either is already set, ensure the value "+
				"is not empty.",
		)
	}

	if secretKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret_key"),
			"Missing AliCloud secret key",
			"The provider cannot create the AliCloud API client as there is a "+
				"missing or empty value for the AliCloud API Secret Key. Set the "+
				"secret key value in the configuration or use the ALICLOUD_SECRET_KEY "+
				"environment variable. If either is already set, ensure the value "+
				"is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	clientCredentialsConfig := &alicloudOpenapiClient.Config{
		RegionId:        &region,
		AccessKeyId:     &accessKey,
		AccessKeySecret: &secretKey,
	}

	// AliCloud Base Client
	baseClientConfig := clientCredentialsConfig
	baseClient, err := alicloudBaseClient.NewClient(baseClientConfig)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create AliCloud Base API Client",
			"An unexpected error occurred when creating the AliCloud Base API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"AliCloud Base Client Error: "+err.Error(),
		)
		return
	}

	// AliCloud CDN Client
	cdnClientConfig := clientCredentialsConfig
	cdnClient, err := alicloudCdnClient.NewClient(cdnClientConfig)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create AliCloud CDN API Client",
			"An unexpected error occurred when creating the AliCloud CDN API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"AliCloud CDN Client Error: "+err.Error(),
		)
		return
	}

	// AliCloud Antiddos Client
	antiddosClientConfig := clientCredentialsConfig
	antiddosClient, err := alicloudAntiddosClient.NewClient(antiddosClientConfig)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create AliCloud Antiddos API Client",
			"An unexpected error occurred when creating the AliCloud Antiddos API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"AliCloud Antiddos Client Error: "+err.Error(),
		)
		return
	}

	// AliCloud SLB Client
	slbClientConfig := clientCredentialsConfig
	slbClient, err := alicloudSlbClient.NewClient(slbClientConfig)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create AliCloud SLB API Client",
			"An unexpected error occurred when creating the AliCloud SLB API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"AliCloud SLB Client Error: "+err.Error(),
		)
		return
	}

	// AliCloud DNS Client
	dnsClientConfig := clientCredentialsConfig
	dnsClient, err := alicloudDnsClient.NewClient(dnsClientConfig)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create AliCloud DNS API Client",
			"An unexpected error occurred when creating the AliCloud DNS API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"AliCloud DNS Client Error: "+err.Error(),
		)
		return
	}

	// AliCloud RAM Client
	ramClientConfig := clientCredentialsConfig
	ramClient, err := alicloudRamClient.NewClient(ramClientConfig)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create AliCloud RAM API Client",
			"An unexpected error occurred when creating the AliCloud RAM API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"AliCloud RAM Client Error: "+err.Error(),
		)
		return
	}

	// AliCloud CMS Client
	cmsClientConfig := clientCredentialsConfig
	cmsClientConfig.Endpoint = tea.String(fmt.Sprintf("metrics.%s.aliyuncs.com", region))
	cmsClient, err := alicloudCmsClient.NewClient(cmsClientConfig)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create AliCloud CMS API Client",
			"An unexpected error occurred when creating the AliCloud CMS API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"AliCloud CMS Client Error: "+err.Error(),
		)
		return
	}

	// AliCloud ADB Client
	adbClientConfig := clientCredentialsConfig
	adbClientConfig.Endpoint = tea.String("adb.aliyuncs.com")
	adbClient, err := alicloudAdbClient.NewClient(adbClientConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create AliCloud ADB API Client",
			"An unexpected error occurred when creating the AliCloud ADB API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"AliCloud ADB Client Error: "+err.Error(),
		)
		return
	}

	// AliCloud EMR Client
	emrClientConfig := clientCredentialsConfig
	emrClientConfig.Endpoint = tea.String(fmt.Sprintf("emr.%s.aliyuncs.com", region))
	emrClient, err := alicloudEmrClient.NewClient(emrClientConfig)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create AliCloud EMR API Client",
			"An unexpected error occurred when creating the AliCloud EMR API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"AliCloud EMR Client Error: "+err.Error(),
		)
		return
	}

	// AliCloud CS Client
	csClientConfig := clientCredentialsConfig
	csClientConfig.Endpoint = tea.String(fmt.Sprintf("cs.%s.aliyuncs.com", region))
	csClient, err := alicloudCsClient.NewClient(csClientConfig)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create AliCloud CS API Client",
			"An unexpected error occurred when creating the AliCloud CS API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"AliCloud CS Client Error: "+err.Error(),
		)
		return
	}

	// AliCloud clients wrapper
	alicloudClients := alicloudClients{
		baseClient:     baseClient,
		cdnClient:      cdnClient,
		antiddosClient: antiddosClient,
		slbClient:      slbClient,
		dnsClient:      dnsClient,
		ramClient:      ramClient,
		cmsClient:      cmsClient,
		adbClient:      adbClient,
		emrClient:      emrClient,
		csClient:       csClient,
	}

	resp.DataSourceData = alicloudClients
	resp.ResourceData = alicloudClients
}

func (p *alicloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCdnDomainDataSource,
		NewDdosCooInstancesDataSource,
		NewDdosCooDomainResourcesDataSource,
		NewSlbLoadBalancersDataSource,
		NewCsUserKubeconfigDataSource,
	}
}

func (p *alicloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAliDnsRecordWeightResource,
		NewAliDnsGtmInstanceResource,
		NewRamUserGroupAttachmentResource,
		NewRamPolicyResource,
		NewCmsAlarmRuleResource,
		NewAlidnsDomainAttachmentResource,
		NewAlidnsInstanceResource,
		NewCmsSystemEventContactGroupAttachmentResource,
		NewDdosCooWebconfigSslAttachmentResource,
		NewAliadbResourceGroupBindResource,
		NewEmrMetricAutoScalingRulesResource,
		NewDdosCooWebAIProtectConfigResource,
	}
}
