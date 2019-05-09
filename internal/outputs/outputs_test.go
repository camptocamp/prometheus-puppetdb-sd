package outputs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	_, err := Setup(&Options{
		Name: "stdout",
	})
	assert.Nil(t, err)
}
