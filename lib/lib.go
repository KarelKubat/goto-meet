// Package lib is a collection of useful stand-alone methods.
package lib

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
)

// ExpandPath is a helper to expand typical Unix shortands in paths.
func ExpandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("cannot find user's homedir: %v", err)
		}
		return filepath.Join(usr.HomeDir, p[2:]), nil
	}
	return p, nil
}
