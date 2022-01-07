package deploy

import "testing"

func TestGenSystemdUnit(t *testing.T) {
	s := &Service{
		Name:      "example",
		NeedState: true,
	}
	t.Logf("%s", s.genSystemdUnit())
}
