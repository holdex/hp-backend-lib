package librollbar

import (
	"os"

	"github.com/rollbar/rollbar-go"
)

func Setup() {
	rollbar.SetToken(os.Getenv("ROLLBAR_TOKEN"))
	rollbar.SetEnvironment(os.Getenv("ROLLBAR_ENVIRONMENT"))
	rollbar.SetCodeVersion(os.Getenv("ROLLBAR_CODE_VERSION"))
	rollbar.SetServerHost(os.Getenv("ROLLBAR_SERVICE"))
}
