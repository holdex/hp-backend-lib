package pg_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"bitbucket.org/holdex/hp-backend-lib/eventstore"
	"bitbucket.org/holdex/hp-backend-lib/eventstore/pg"
	"bitbucket.org/holdex/hp-backend-lib/log"
)

var (
	texts = map[uint64]string{
		32:    "qwertyqwertyuiopasdfghjklzxcvbnm",
		64:    "qwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnm",
		128:   "qwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnm",
		256:   "qwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnm",
		512:   "qwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnm",
		1024:  "qwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnm",
		2048:  "qwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnm",
		4096:  "qwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnm",
		8192:  "qwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnm",
		16384: "qwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnm",
	}

	sqlDN  = "postgres"
	sqlDSN = "postgresql://USERNAME:PASSWORD@HOST/DATABASE?sslmode=disable"
)

func BenchmarkService_StoreEvents(b *testing.B) {
	db, err := sql.Open(sqlDN, sqlDSN)
	if err != nil {
		liblog.Fatalf("failed to connect to postgres [SQL_DRIVER_NAME=%s] [SQL_DATA_SOURCE_NAME=%s]: %v", sqlDN, sqlDSN, err)
	}
	defer db.Close()

	s := pg.NewService(db)

	store := func(events uint64, payload string) error {
		var eventsToStore []libeventstore.Event
		streamID := uuid.NewV4().String()
		for j := uint64(1); j <= events; j++ {
			eventsToStore = append(eventsToStore, libeventstore.Event{
				StreamId:       streamID,
				StreamType:     "user_account",
				StreamRevision: j,
				Type:           fmt.Sprintf("%duser_account.CreatedWithTelegram", j),
				Payload:        []byte(payload),
				CreatedAt:      int64(1540910403871836713),
				Metadata:       []byte("qwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnm"),
			})
		}
		return s.StoreEvents(context.Background(), eventsToStore...)
	}

	for events := uint64(1); events <= 32; events++ {
		for chars := uint64(32); chars <= 16384; chars *= 2 {
			text := texts[chars]
			b.Run(fmt.Sprintf("%d events %d chars", events, chars), func(b *testing.B) {
				for bi := 0; bi < b.N; bi++ {
					assert.Nil(b, store(events, text))
				}
			})
		}
	}
}

func TestService_StoreEvents(t *testing.T) {
	db, err := sql.Open(sqlDN, sqlDSN)
	if err != nil {
		liblog.Fatalf("failed to connect to postgres [SQL_DRIVER_NAME=%s] [SQL_DATA_SOURCE_NAME=%s]: %v", sqlDN, sqlDSN, err)
	}
	defer db.Close()

	s := pg.NewService(db)

	store := func(events uint64, payload string) error {
		var eventsToStore []libeventstore.Event
		streamID := uuid.NewV4().String()
		for j := uint64(1); j <= events; j++ {
			eventsToStore = append(eventsToStore, libeventstore.Event{
				StreamId:       streamID,
				StreamType:     "user_account",
				StreamRevision: j,
				Type:           fmt.Sprintf("%duser_account.CreatedWithTelegram", j),
				Payload:        []byte(payload),
				CreatedAt:      int64(1540910403871836713),
				Metadata:       []byte("qwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnmqwertyqwertyuiopasdfghjklzxcvbnm"),
			})
		}
		return s.StoreEvents(context.Background(), eventsToStore...)
	}

	for events := uint64(1); events <= 32; events++ {
		for chars := uint64(32); chars <= 16384; chars *= 2 {
			text := texts[chars]
			t.Run(fmt.Sprintf("%d events %d chars", events, chars), func(b *testing.T) {
				assert.Nil(b, store(events, text))
			})
		}
	}
}

func BenchmarkService_StreamEvents(b *testing.B) {
	db, err := sql.Open(sqlDN, sqlDSN)
	if err != nil {
		liblog.Fatalf("failed to connect to postgres [SQL_DRIVER_NAME=%s] [SQL_DATA_SOURCE_NAME=%s]: %v", sqlDN, sqlDSN, err)
	}
	defer db.Close()

	s := pg.NewService(db)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b.ResetTimer()
	stream := s.StreamEvents(ctx, 0, 100, "1user_account.CreatedWithTelegram",
		"2user_account.CreatedWithTelegram",
		"3user_account.CreatedWithTelegram",
		"4user_account.CreatedWithTelegram",
		"5user_account.CreatedWithTelegram",
		"6user_account.CreatedWithTelegram",
		"7user_account.CreatedWithTelegram",
		"8user_account.CreatedWithTelegram",
		"9user_account.CreatedWithTelegram",
		"10user_account.CreatedWithTelegram")
	if err != nil {
		assert.Fail(b, "failed to stream: %v", err)
	}

	go func() {
		time.Sleep(10 * time.Second)
		cancel()
	}()

	b.N = 0
	for {
		select {
		case <-ctx.Done():
			b.StopTimer()
			return
		case streamEvent := <-stream:
			if streamEvent.Err != nil {
				b.Fatalf("received error: %v", streamEvent.Err)
			}
			b.N++
			<-time.After(10 * time.Millisecond) // Simulate processing load
		}
	}
}
