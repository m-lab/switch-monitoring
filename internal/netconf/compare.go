package netconf

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kylelemons/godebug/diff"
)

// Compare cleans up two switch configuration files and returns true if they
// are the same.
func Compare(c1, c2 string) bool {
	c1 = cleanConfig(c1)
	c2 = cleanConfig(c2)

	fmt.Println(diff.Diff(c1, c2))
	return c1 == c2
}

// cleanConfig cleans a JunOS switch configuration file to make it comparable.
// It removes any comment lines at the beginning, trims whitespace and the
// beginning/end, replaces any encrypted password with "dummy" and returns
// what is left.
func cleanConfig(config string) string {
	// Remove version string. This is necessary since the switch will always
	// report the current version number at the beginning of the config file,
	// but the version is not part of the config file itself.
	re := regexp.MustCompile("version.+")
	config = re.ReplaceAllString(config, "")

	// Remove comments (lines starting with '#').
	re = regexp.MustCompile("(?m)^#.*$")
	config = strings.TrimSpace(re.ReplaceAllString(config, ""))

	// Replace all password fields with "dummy".
	// TODO: once we start pre-configuring the switch with a random password,
	// we can remove this step so that the actual passwords are compared.
	re = regexp.MustCompile("encrypted-password.+")
	config = re.ReplaceAllString(config, "encrypted-password \"dummy\";")

	// Trim the result.
	return strings.TrimSpace(config)
}
