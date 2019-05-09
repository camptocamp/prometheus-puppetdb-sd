package outputs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	o, err := Setup(&Options{
		Name: "stdout",
	})
	assert.Nil(t, err)
	assert.Equal(t, "stdout", o.GetName())
}
