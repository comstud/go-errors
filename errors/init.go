package errors

import (
	"strings"

	"github.com/comstud/go-rollbar/rollbar"
	"github.com/pborman/uuid"
)

type ErrIDGenerator interface {
	GenErrID() string
}

type config struct {
	ErrIDGenerator ErrIDGenerator
	RollbarClient  *rollbar.Client
}

type defaultErrIDGenerator struct{}

func (defaultErrIDGenerator) GenErrID() string {
	return "ERR" + strings.ToUpper(
		strings.Replace(
			uuid.New(),
			"-",
			"",
			-1,
		),
	)
}

var RegisteredErrors = make(map[string]*ErrorClass)
var Config *config

func init() {
	Config = &config{
		ErrIDGenerator: &defaultErrIDGenerator{},
	}
}
