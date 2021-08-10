package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/utilitywarehouse/terraform-provider-nebraska/nebraska"
)

func resourceChannel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceChannelCreate,
		ReadContext:   resourceChannelRead,
		UpdateContext: resourceChannelUpdate,
		DeleteContext: resourceChannelDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"arch": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"all", "amd64", "aarch64", "x86"}, false),
			},
			"application_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"color": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"created_ts": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"package_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceChannelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient)

	appID, err := getApplicationID(d, c)
	if err != nil {
		return diag.FromErr(err)
	}

	arch, err := nebraska.ArchFromString(d.Get("arch").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	input := &nebraska.AddChannelInput{
		Name:      d.Get("name").(string),
		Color:     d.Get("color").(string),
		PackageID: d.Get("package_id").(string),
		Arch:      arch,
	}

	channel, err := c.AddChannel(appID, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(channel.ID)

	return resourceChannelRead(ctx, d, meta)
}

func resourceChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient)

	appID, err := getApplicationID(d, c)
	if err != nil {
		return diag.FromErr(err)
	}
	channel, err := c.GetChannel(appID, d.Id())
	if err != nil {
		if err == nebraska.ErrNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if channel == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("name", channel.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("color", channel.Color); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("created_ts", channel.CreatedTs.String()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("package_id", channel.PackageID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceChannelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient)

	appID, err := getApplicationID(d, c)
	if err != nil {
		return diag.FromErr(err)
	}

	input := &nebraska.UpdateChannelInput{
		Name:      d.Get("name").(string),
		Color:     d.Get("color").(string),
		PackageID: d.Get("package_id").(string),
	}

	if _, err := c.UpdateChannel(appID, d.Id(), input); err != nil {
		return diag.FromErr(err)
	}

	return resourceChannelRead(ctx, d, meta)
}

func resourceChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient)

	appID, err := getApplicationID(d, c)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := c.DeleteChannel(appID, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
