package tablo_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vigo/tablo/internal/tablo"
)

type BytesWriteCloser struct {
	bytes.Buffer
}

func (s *BytesWriteCloser) Close() error {
	return nil
}

func TestTablo_New_with_no_args(t *testing.T) {
	tbl, err := tablo.New()

	assert.NotNil(t, tbl)
	assert.NoError(t, err)

	assert.Equal(t, tablo.Version, tbl.Version)
	assert.Equal(t, os.Stdout, tbl.Output)
	assert.NotNil(t, tbl.ReadInputFunc)
	assert.NotNil(t, tbl.ParseArgsFunc)
	assert.Empty(t, tbl.Args)
	assert.Equal(t, int32(0), tbl.LineDelimeter)
	assert.Equal(t, int32(0), tbl.FieldDelimeter)
	assert.False(t, tbl.DisplayVersion)
	assert.False(t, tbl.SeparateRows)
}

func TestTablo_New_with_nil_ParseArgsFunc(t *testing.T) {
	tbl, err := tablo.New(
		tablo.WithParseArgsFunc(nil),
	)

	assert.Nil(t, tbl)
	assert.Error(t, err)
	assert.ErrorIs(t, err, tablo.ErrValueRequired)
}

func TestTablo_New_with_nil_ReadInputFunc(t *testing.T) {
	tbl, err := tablo.New(
		tablo.WithReadInputFunc(nil),
	)

	assert.Nil(t, tbl)
	assert.Error(t, err)
	assert.ErrorIs(t, err, tablo.ErrValueRequired)
}

func TestTablo_New_WithArgs(t *testing.T) {
	tbl, err := tablo.New(
		tablo.WithArgs([]string{"FOO"}),
	)

	assert.NotNil(t, tbl)
	assert.NoError(t, err)
	assert.Equal(t, []string{"FOO"}, tbl.Args)
}

func TestTablo_New_WithOutput(t *testing.T) {
	tbl, err := tablo.New(
		tablo.WithOutput("stdout"),
	)

	assert.NotNil(t, tbl)
	assert.NoError(t, err)
	assert.Equal(t, os.Stdout, tbl.Output)
}

func TestTablo_New_WithDisplayVersion(t *testing.T) {
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stderr = w

	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithDisplayVersion(true),
	)

	assert.NotNil(t, tbl)
	assert.NoError(t, err)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	assert.Equal(t, tablo.Version, string(bytes.TrimSpace(buf.Bytes())))
}
