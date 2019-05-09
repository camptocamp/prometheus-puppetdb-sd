package outputs

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStdoutWriteOutputSuccess(t *testing.T) {
	data := map[string]string{
		"foo": "bar",
	}
	o := &OutputStdout{}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := o.WriteOutput(data)

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = old
	out := <-outC

	assert.Nil(t, err)
	assert.Equal(t, "foo: bar\n", out)
}
