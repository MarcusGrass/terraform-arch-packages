package pacman

import (
	"fmt"
	"testing"
)

func TestInstallUninstall(t *testing.T) {
	pkg := "zip"
	err := installPackage(pkg, "")
	if err != nil {
		t.Fatalf("Failed to install zip pkg")
	}
	if ok, _ := packageIsInstalled(pkg); !ok {
		t.Fatalf("Package was not confirmed installed")
	}
	err = uninstallPackage(pkg, true)
	if err != nil {
		t.Fatalf("Failed to delete zip pkg")
	}
	if ok, _ := packageIsInstalled(pkg); ok {
		t.Fatalf("Package was not deleted from system")
	}
	err = installPackage(pkg, "")
	if err != nil {
		t.Fatalf("Failed to delete zip pkg")
	}
	if ok, _ := packageIsInstalled(pkg); !ok {
		t.Fatalf("Package was not confirmed installed after reinstall")
	}
}

func TestErr(t *testing.T) {
	err := installPackage("abcd1234", "")
	fmt.Println(err)
}
