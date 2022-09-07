package core

import (
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core/database"
	"time"
)

type JPNICCert struct {
	Base      database.JPNIC
	RenewDate time.Time
}
