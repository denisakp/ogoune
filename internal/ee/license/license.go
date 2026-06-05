// SPDX-License-Identifier: LicenseRef-Ogoune-EE
package license

import (
	"os"
	"strings"
)

type Edition string

const (
	Community  Edition = "community"
	Enterprise Edition = "enterprise"
)

const enterprisePrefix = "pg_ent_"

func Get() Edition {
	key := os.Getenv("ENTERPRISE_LICENSE_KEY")
	if strings.HasPrefix(key, enterprisePrefix) {
		return Enterprise
	}
	return Community
}

func IsEnterprise() bool {
	return Get() == Enterprise
}

// PoweredByRequired reports whether the public status page must display the
// "Powered by Ogoune" attribution. Community edition always requires it;
// Enterprise may suppress it.
func PoweredByRequired() bool {
	return !IsEnterprise()
}
