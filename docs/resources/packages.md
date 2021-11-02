# Resource `pacman_packages`

The pacman packages resource allows you to specify packages by name and delete behaviour

## Example Usage

```terraform
resource "pacman_packages" "base" {
  cascade_on_delete = true,
  packages {
    name = "alacritty"
  }
  packages {
    name = "zip"
  }
}
```

## Argument Reference

- `packages` - (Required) All (explicitly installed) that should be present on the system.
- `cascade_on_delete` - (Optional) Use pacman -Rs instead of pacman -R on uninstall.
- `sudo_password` - (Optional) Required if not root or user has nopasswd sudo, will be piped to stdin for install/uninstall cmds.

### Packages

Each package contains the package name (that will be run as pacman -S <name>)

- `name` - (Required) package name (that will be run as pacman -S <name>).
