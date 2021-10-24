package pacman

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func pacmanPackages() *schema.Resource {
	return &schema.Resource{
		ReadContext: packageRead,
		Schema: map[string]*schema.Schema{
			"packages": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func packageRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	manualInstalled, err := findExplicitlyInstalledPackages()

	if err != nil {
		return diag.FromErr(err)
	}

	packages := make([]map[string]string, 0)

	for _, pkg := range manualInstalled {
		packages = append(packages, map[string]string{
			"name":    pkg.name,
			"version": pkg.version,
		})
	}

	if err := d.Set("packages", packages); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
