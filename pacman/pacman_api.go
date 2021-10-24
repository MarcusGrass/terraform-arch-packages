package pacman

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

type PackageData struct {
	name              string
	version           string
	hasDependent      bool
	installedByPacman bool
}

func installPackage(name string, pw string) error {
	cmd := WithSudoIfNotRoot(pw, "pacman", "-S", name, "--noconfirm")
	_, err := execConvertErr(cmd)
	return err
}

func packageIsInstalled(name string) (bool, error) {
	_, err := execConvertErr(exec.Command("pacman", "-Qi", name))
	if err != nil {
		msg := err.Error()
		errmsg := fmt.Sprintf("error: package '%v' was not found", name)
		if strings.HasPrefix(msg, errmsg) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, err
}

func uninstallPackage(name string, deleteDeps bool, pw string) error {
	if deleteDeps {
		_, err := execConvertErr(WithSudoIfNotRoot(pw, "pacman", "-Rs", "--noconfirm", name))
		return err
	} else {
		_, err := execConvertErr(WithSudoIfNotRoot(pw, "pacman", "-R", "--noconfirm", name))
		return err
	}
}

func findExplicitlyInstalledPackages() (map[string]PackageData, error) {
	all, err := getPackageData(false)
	if err != nil {
		return nil, err
	}
	return all, nil
}

func getPackageData(includeForeign bool) (map[string]PackageData, error) {
	pacSearch := "-Qni"
	if includeForeign {
		pacSearch = "-Qmi"
	}
	cmd := exec.Command("pacman", pacSearch)
	stdout, e := execConvertErr(cmd)
	if e != nil {
		return nil, e
	}
	pkgData := make(map[string]PackageData, 0)
	for _, row := range strings.Split(stdout, "\n\n") {
		data := strings.Split(row, "\n")
		name := ""
		version := ""
		hasDep := true
		wasDeliberatelyInstalled := false
		for _, property := range data {
			if strings.HasPrefix(property, "Name") {
				name = strings.Split(property, ":")[1]
				name = strings.TrimSpace(name)
			}
			if strings.HasPrefix(property, "Version") {
				version = strings.Split(property, ":")[1]
				version = strings.TrimSpace(version)
			}
			if strings.HasPrefix(property, "Required By") {
				req := strings.Split(property, ":")[1]
				req = strings.TrimSpace(req)
				hasDep = req != "None"
			}
			if strings.HasPrefix(property, "Install Reason") {
				req := strings.Split(property, ":")[1]
				req = strings.TrimSpace(req)
				wasDeliberatelyInstalled = req == "Explicitly installed"
			}
		}
		if wasDeliberatelyInstalled {
			pkgData[name] = PackageData{
				name:              name,
				version:           version,
				hasDependent:      hasDep,
				installedByPacman: !includeForeign,
			}
		}
	}
	return pkgData, nil
}

func WithSudoIfNotRoot(sudoPw string, bin string, args ...string) *exec.Cmd {
	if len(sudoPw) > 0 {
		buf := bytes.Buffer{}
		buf.Write([]byte(sudoPw + "\n"))
		cmd := exec.Command("sudo", append([]string{"-S", bin}, args...)...)
		cmd.Stdin = &buf
		return cmd
	} else if u, err := user.Current(); err == nil && u.Username == "root" {
		return exec.Command(bin, args...)
	} else {
		return exec.Command("sudo", append([]string{bin}, args...)...)
	}
}

func execConvertErr(cmd *exec.Cmd) (string, error) {
	r, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	cmd.Stderr = cmd.Stdout
	scanner := bufio.NewScanner(r)
	done := make(chan string)
	go func() {
		out := ""
		for scanner.Scan() {
			line := scanner.Text()
			out += line + "\n"
		}
		done <- out
	}()
	err = cmd.Start()
	if err != nil {
		return "", err
	}

	out := <-done
	err = cmd.Wait()
	if err != nil {
		return "", errors.New("Pacman error: " + out)
	}
	return out, nil
}
