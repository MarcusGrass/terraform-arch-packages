package pacman

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
)

type Package struct {
	name      string
	installed bool
}

const packagesKey = "packages"
const cascadeOnDeleteKey = "cascade_on_delete"
const sudoPasswdKey = "sudo_password"
const installedKey = "installed"
const nameKey = "name"

func resourcePackage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePackageCreate,
		ReadContext:   resourcePackageRead,
		UpdateContext: resourcePackageUpdate,
		DeleteContext: resourcePackageDelete,
		Schema: map[string]*schema.Schema{
			cascadeOnDeleteKey: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Uses pacman -Rs instead of pacman -R to delete packages.",
			},
			sudoPasswdKey: {
				Type:         schema.TypeString,
				Optional:     true,
				InputDefault: "",
				Sensitive:    true,
				Description: "Password for sudo, will be piped into stdin if supplied, only way to run without this is" +
					" if the user is nopasswd sudo or root.",
			},
			packagesKey: {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						nameKey: {
							Type:     schema.TypeString,
							Required: true,
						},
						installedKey: {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourcePackageCreate(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pwd := getSudoPasswd(d)
	statePackages := getPackages(d).List()

	installedPackages, err := findExplicitlyInstalledPackages()
	if err != nil {
		return diag.FromErr(err)
	}
	for _, pkg := range statePackages {
		pkg := pkg.(map[string]interface{})
		name := pkg[nameKey].(string)
		if p, ok := installedPackages[name]; ok {
			if !p.installedByPacman {
				return foreignPackageErr(name)
			}
		} else {
			err := installPackage(name, pwd)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	d.SetId(strconv.FormatInt(time.Now().UnixMilli(), 10))
	return diags
}

func resourcePackageRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pkgs, err := findExplicitlyInstalledPackages()
	if err != nil {
		return diag.FromErr(err)
	}
	installed := make([]interface{}, len(pkgs))
	for _, v := range pkgs {
		installed = append(installed, flattenOne(v))
	}
	cur := d.Get(packagesKey).(*schema.Set)
	if err := d.Set(packagesKey, schema.NewSet(cur.F, installed)); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourcePackageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pkgs, err := findExplicitlyInstalledPackages()
	if err != nil {
		return diag.FromErr(err)
	}
	desiredState := resourceToPackageMap(getPackages(d).List())
	cascade := shouldCascadeOnDelete(d)
	pwd := getSudoPasswd(d)
	for k, v := range pkgs {
		if _, ok := desiredState[k]; !ok {
			if v.installedByPacman {
				err := uninstallPackage(k, cascade, pwd)
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}
	for k := range desiredState {
		if _, ok := pkgs[k]; !ok {
			err := installPackage(k, pwd)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return resourcePackageRead(ctx, d, m)
}

func resourcePackageDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cascade := shouldCascadeOnDelete(d)
	pwd := getSudoPasswd(d)
	for _, packageName := range getPackages(d).List() {
		pkg := packageListItemToPackage(packageName)
		if err := uninstallPackage(pkg.name, cascade, pwd); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return diags
}

func flattenOne(packages PackageData) interface{} {
	p := make(map[string]interface{})
	p[nameKey] = packages.name
	p[installedKey] = true
	return p
}

func resourceToPackageMap(state []interface{}) map[string]Package {
	m := make(map[string]Package)
	for _, val := range state {
		pkg := packageListItemToPackage(val)
		m[pkg.name] = pkg
	}
	return m
}

func packageListItemToPackage(pkg interface{}) Package {
	p := pkg.(map[string]interface{})
	return Package{
		name:      p[nameKey].(string),
		installed: p[installedKey].(bool),
	}
}
func getPackages(d *schema.ResourceData) *schema.Set {
	return d.Get(packagesKey).(*schema.Set)
}

func shouldCascadeOnDelete(d *schema.ResourceData) bool {
	return d.Get(cascadeOnDeleteKey).(bool)
}

func getSudoPasswd(d *schema.ResourceData) string {
	return d.Get(sudoPasswdKey).(string)
}

func foreignPackageErr(name string) diag.Diagnostics {
	return diag.FromErr(errors.New(fmt.Sprintf("Package <%v> not installed by pacman, remove it from packages and try again", name)))
}
