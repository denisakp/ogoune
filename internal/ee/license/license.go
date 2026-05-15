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
