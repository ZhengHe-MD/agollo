package properties

import (
	"testing"

	. "github.com/tevid/gohamcrest"
)

var (
	propertiesParser = &Parser{}
)

func TestPropertiesParser(t *testing.T) {
	s, err := propertiesParser.Parse(`test`)
	Assert(t, err, NilVal())
	Assert(t, s, NilVal())
}
