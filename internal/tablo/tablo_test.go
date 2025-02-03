package tablo_test

import (
	"bytes"
	"io"
	"os"
	"strings"
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

func TestMain_Run_PipedInput(t *testing.T) {
	input := bytes.NewBufferString("hello world\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimeter(" "),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			b := input.Bytes()
			return string(b[:len(b)-1]), nil
		}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	// Expected table output
	expectedOutput := `┌─────────────┐
│ hello world │
└─────────────┘
`
	lines := strings.Split(output.String(), "\n")
	assert.Equal(t, expectedOutput, strings.Join(lines[1:], "\n"))
}

func TestMain_Run_PipedInput_with_delimeter(t *testing.T) {
	input := bytes.NewBufferString("hello:world\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimeter(":"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			b := input.Bytes()
			return string(b[:len(b)-1]), nil
		}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `┌───────┬───────┐
│ hello │ world │
└───────┴───────┘
`
	lines := strings.Split(output.String(), "\n")
	assert.Equal(t, expectedOutput, strings.Join(lines[1:], "\n"))
}

func TestMain_Run_PipedInput_docker_images(t *testing.T) {
	input := bytes.NewBufferString(`REPOSITORY                         TAG       IMAGE ID       CREATED         SIZE
superset-superset-worker-beat      latest    3292fc2e6758   2 weeks ago     958MB
superset-superset-worker           latest    d25cbcc60691   2 weeks ago     958MB
`)
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimeter(" "),
		tablo.WithLineDelimeter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			b := input.Bytes()
			return string(b[:len(b)-1]), nil
		}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `┌───────────────────────────────┬────────┬──────────────┬─────────────┬───────┐
│ REPOSITORY                    │ TAG    │ IMAGE ID     │ CREATED     │ SIZE  │
├───────────────────────────────┼────────┼──────────────┼─────────────┼───────┤
│ superset-superset-worker-beat │ latest │ 3292fc2e6758 │ 2 weeks ago │ 958MB │
├───────────────────────────────┼────────┼──────────────┼─────────────┼───────┤
│ superset-superset-worker      │ latest │ d25cbcc60691 │ 2 weeks ago │ 958MB │
└───────────────────────────────┴────────┴──────────────┴─────────────┴───────┘
`
	lines := strings.Split(output.String(), "\n")
	assert.Equal(t, expectedOutput, strings.Join(lines[1:], "\n"))
}

func TestMain_Run_PipedInput_docker_images_with_args(t *testing.T) {
	input := bytes.NewBufferString(`REPOSITORY                         TAG       IMAGE ID       CREATED         SIZE
superset-superset-worker-beat      latest    3292fc2e6758   2 weeks ago     958MB
superset-superset-worker           latest    d25cbcc60691   2 weeks ago     958MB
`)
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimeter(" "),
		tablo.WithLineDelimeter("\n"),
		tablo.WithArgs([]string{"REPOSITORY"}),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			b := input.Bytes()
			return string(b[:len(b)-1]), nil
		}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `┌───────────────────────────────┐
│ REPOSITORY                    │
├───────────────────────────────┤
│ superset-superset-worker-beat │
├───────────────────────────────┤
│ superset-superset-worker      │
└───────────────────────────────┘
`
	lines := strings.Split(output.String(), "\n")
	assert.Equal(t, expectedOutput, strings.Join(lines[1:], "\n"))
}

func TestMain_Run_PipedInput_docker_images_with_no_separation(t *testing.T) {
	input := bytes.NewBufferString(`REPOSITORY                         TAG       IMAGE ID       CREATED         SIZE
superset-superset-worker-beat      latest    3292fc2e6758   2 weeks ago     958MB
superset-superset-worker           latest    d25cbcc60691   2 weeks ago     958MB
`)
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimeter(" "),
		tablo.WithLineDelimeter("\n"),
		tablo.WithSeparateRows(true),
		tablo.WithArgs([]string{"REPOSITORY"}),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			b := input.Bytes()
			return string(b[:len(b)-1]), nil
		}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `┌───────────────────────────────┐
│ REPOSITORY                    │
├───────────────────────────────┤
│ superset-superset-worker-beat │
│ superset-superset-worker      │
└───────────────────────────────┘
`
	lines := strings.Split(output.String(), "\n")
	assert.Equal(t, expectedOutput, strings.Join(lines[1:], "\n"))
}

func TestRun_NoArgs(t *testing.T) {
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
	)
	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "press ENTER")
}
