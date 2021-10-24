package pacman

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"pacman_packages": resourcePackage(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"pacman_packages": pacmanPackages(),
		},
	}
}
