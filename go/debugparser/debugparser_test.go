package debugparser_test

import (
	"testing"

	"go.astrophena.name/exp/go/debugparser"
)

func TestFetch(t *testing.T) {
	d, err := debugparser.Fetch("go.astrophena.name")
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", d)
}
