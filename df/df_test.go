package df

import (
	"os/exec"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	out, err := exec.Command("df", "-h").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(out), "\n")
	var dfe [][]string
	for _, line := range lines[1 : len(lines)-1] {
		dfe = append(dfe, strings.Fields(line))
	}
	t.Logf("%+v", dfe)
}
