# A provider for pacman packages
Manages installation of pacman packages, theoretically works on most places where pacman is installed.
## Danger notice
While care has been put into making usage clear, this is a roundabout way of running shell commands that require root permissions.
This has some implications on the security of the system while also presenting a risk of accidentally deleting packages. 
Automatically installing/uninstalling packages without confirming always poses a risk which users should be aware of and understand.

### Imperfect implementation details:
* Uses golang to run pacman commands through an os/exec command
* Requires root user or user with sudo, see parameters for details

### Caveats
The provider only manipulates explicitly installed packages (`pacman -Qi`-entry contains a row with "Install Reason  : Explicitly installed").  
While being able to list/remove any package might be desirable on minimal systems it would be difficult to maintain 
such a configuration and increase the likelihood of breaking the system.

### Features
* On create will ensure the packages are installed by installing if missing. Will not remove existing packages.
* On update will uninstall packages that are unlisted, and install added packages
* On delete will uninstall previously listed packages.

### Parameters:
* *sudo_password*: Required if not root or user has nopasswd sudo, will be piped to stdin for install/uninstall cmds
* *cascade_on_delete*: Default false uses `pacman -R <pkg> --noconfirm`, true switches it to `pacman -Rs <pkg> --noconfirm`
* *packages*: A list of packages with a name property (`pacman -S <name> --noconfirm`)
