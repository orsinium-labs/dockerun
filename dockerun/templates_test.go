package dockerun

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplatesName(t *testing.T) {
	is := require.New(t)
	is.Equal(tName(""), "")
	is.Equal(tName("hello"), "hello")
	is.Equal(tName("hello==13.1.0"), "hello")
	is.Equal(tName("github.com/stretchr/testify"), "testify")
	is.Equal(tName("github.com/stretchr/testify/"), "testify")
	is.Equal(tName("https://github.com/stretchr/testify/"), "testify")
}
