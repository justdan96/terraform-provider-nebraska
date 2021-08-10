package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/utilitywarehouse/terraform-provider-nebraska/nebraska"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in
	// document generation and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output.  Add defaults on
	// to the exported descriptions if present.
	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		return strings.TrimSpace(desc)
	}
}

// New returns a new provider
func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"application_id": {
					Type:         schema.TypeString,
					Optional:     true,
					DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"NEBRASKA_APPLICATION_ID"}, ""),
					ValidateFunc: validation.StringIsNotWhiteSpace,
				},
				"endpoint": {
					Type:         schema.TypeString,
					Optional:     true,
					DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"NEBRASKA_ENDPOINT"}, "http://localhost:8000"),
					ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"nebraska_channel": dataSourceChannel(),
				"nebraska_group":   dataSourceGroup(),
				"nebraska_package": dataSourcePackage(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"nebraska_channel": resourceChannel(),
				"nebraska_group":   resourceGroup(),
				"nebraska_package": resourcePackage(),
			},
		}

		p.ConfigureContextFunc = providerConfigure(version, p)

		return p
	}
}

type apiClient struct {
	*nebraska.Client
	ApplicationID string
}

func providerConfigure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		c := nebraska.New(d.Get("endpoint").(string), p.UserAgent("terraform-provider-nebraska", version))
		return &apiClient{
			Client:        c,
			ApplicationID: d.Get("application_id").(string),
		}, nil
	}
}

func getApplicationID(d *schema.ResourceData, client *apiClient) (string, error) {
	if id, ok := d.GetOk("application_id"); ok {
		return id.(string), nil
	}
	if client.ApplicationID != "" {
		return client.ApplicationID, nil
	}

	return "", fmt.Errorf("application_id: required field is not set")
}
