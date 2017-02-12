package tests

import (
	"io/ioutil"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleUsage(t *testing.T) {
	valuesFile := `
foo: bar
key: value
`
	inputFile := "foo is {{ .foo }} and keyx2 is {{ .key | repeat 2 }}"
	f, err := ioutil.TempFile("", "sprig_testval")
	assert.Nil(t, err)
	f.WriteString(valuesFile)
	f2, err := ioutil.TempFile("", "sprig_testin")
	f2.WriteString(inputFile)

	cmd := exec.Command("../bin/sprig", "-f", f.Name(), f2.Name())
	out, err := cmd.CombinedOutput()
	assert.Nil(t, err)
	assert.Equal(t, string(out), "foo is bar and keyx2 is valuevalue")
}
