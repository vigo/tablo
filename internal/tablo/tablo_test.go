package tablo_test

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"io/fs"
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

func (s *BytesWriteCloser) chomp() []byte {
	return bytes.TrimSuffix(s.Bytes(), []byte("\n"))
}

func (s *BytesWriteCloser) nonStdinValue() []byte {
	firstLineBreakIndex := bytes.IndexByte(s.Bytes(), '\n')
	if firstLineBreakIndex > 0 {
		return s.Bytes()[firstLineBreakIndex+1:]
	}

	return s.Bytes()
}

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestTablo_New_No_Options(t *testing.T) {
	tbl, err := tablo.New()

	assert.NotNil(t, tbl)
	assert.NoError(t, err)
}

func TestTablo_New_WithArgs_Nil(t *testing.T) {
	tbl, err := tablo.New(
		tablo.WithArgs(nil),
	)

	assert.NotNil(t, tbl)
	assert.NoError(t, err)
	assert.Equal(t, []string(nil), tbl.Args)
}

func TestTablo_New_WithOutputWriter_Nil(t *testing.T) {
	tbl, err := tablo.New(
		tablo.WithOutputWriter(nil),
	)

	assert.Nil(t, tbl)
	assert.Error(t, err)
	assert.ErrorIs(t, err, tablo.ErrValueRequired)
}

func TestTablo_New_WithOutput_EmptyString(t *testing.T) {
	tbl, err := tablo.New(
		tablo.WithOutput(""),
	)

	assert.Nil(t, tbl)
	assert.Error(t, err)
	assert.ErrorIs(t, err, tablo.ErrValueRequired)
}

func TestTablo_New_WithOutput_InvalidFilePath(t *testing.T) {
	tbl, err := tablo.New(
		tablo.WithOutput("/<>/()/"),
	)

	assert.Nil(t, tbl)
	assert.Error(t, err)
	assert.ErrorIs(t, err, tablo.ErrInvalidValue)
}

func TestTablo_New_WithOutput_ValidFile(t *testing.T) {
	outputFile, err := os.CreateTemp("", "output.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(outputFile.Name()) }()

	tbl, err := tablo.New(
		tablo.WithOutput(outputFile.Name()),
	)

	assert.NotNil(t, tbl)
	assert.NoError(t, err)
	assert.Implements(t, (*io.Writer)(nil), tbl.Output)
}

func TestTablo_New_WithReadInputFunc_Nil(t *testing.T) {
	tbl, err := tablo.New(
		tablo.WithReadInputFunc(nil),
	)

	assert.Nil(t, tbl)
	assert.Error(t, err)
	assert.ErrorIs(t, err, tablo.ErrValueRequired)
}

func TestTablo_New_WithLineDelimiter_EmptyString(t *testing.T) {
	tbl, err := tablo.New(
		tablo.WithLineDelimiter(""),
	)

	assert.Nil(t, tbl)
	assert.Error(t, err)
	assert.ErrorIs(t, err, tablo.ErrValueRequired)
}

func TestTablo_New_WithFieldDelimiter_EmptyString(t *testing.T) {
	tbl, err := tablo.New(
		tablo.WithFieldDelimiter(""),
	)

	assert.NotNil(t, tbl)
	assert.NoError(t, err)
}

func TestTablo_GetVersion(t *testing.T) {
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stderr = w

	os.Args = []string{"tablo", "-version"}
	resetFlags()

	output := new(BytesWriteCloser)

	err = tablo.Run()
	assert.NoError(t, err)

	_ = w.Close()
	os.Stdout = oldStdout

	_, _ = output.ReadFrom(r)

	assert.Equal(t, tablo.Version, string(output.chomp()))
}

func TestTablo_Tabelize_SingleString(t *testing.T) {
	input := bytes.NewBufferString("hello world\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithLineDelimiter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := "┌─────────────┐\n│ hello world │\n└─────────────┘\n"
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_SingleString_With_TAB_LineDelimiter(t *testing.T) {
	input := bytes.NewBufferString("hello\tworld")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithLineDelimiter("\t"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)
	expectedOutput := "┌───────┐\n│ hello │\n├───────┤\n│ world │\n└───────┘\n"
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_SingleString_With_R_LineDelimiter(t *testing.T) {
	input := bytes.NewBufferString("hello\rworld")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithLineDelimiter("\r"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)
	expectedOutput := "┌───────┐\n│ hello │\n├───────┤\n│ world │\n└───────┘\n"
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_SingleString_No_Borders(t *testing.T) {
	input := bytes.NewBufferString("hello\nworld")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithNoDrawBorder(true),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)
	expectedOutput := " hello \n world \n"
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_WithFieldDelimiter(t *testing.T) {
	input := bytes.NewBufferString("hello1|world1\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("|"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := "┌────────┬────────┐\n│ hello1 │ world1 │\n└────────┴────────┘\n"
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_WithFieldDelimiter_TAB(t *testing.T) {
	input := bytes.NewBufferString("hello2\tworld2\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("\t"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := "┌────────┬────────┐\n│ hello2 │ world2 │\n└────────┴────────┘\n"
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_WithFieldDelimiter_VERTICAL_TAB(t *testing.T) {
	input := bytes.NewBufferString("hello3\vworld3\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("\v"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := "┌────────┬────────┐\n│ hello3 │ world3 │\n└────────┴────────┘\n"
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_WithFieldDelimiter_BACKSPACE(t *testing.T) {
	input := bytes.NewBufferString("hello4\bworld4\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("\b"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := "┌────────┬────────┐\n│ hello4 │ world4 │\n└────────┴────────┘\n"
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_WithFieldDelimiter_BELL(t *testing.T) {
	input := bytes.NewBufferString("hello5\aworld5\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("\a"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := "┌────────┬────────┐\n│ hello5 │ world5 │\n└────────┴────────┘\n"
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_WithFieldDelimiter_FORM_FIELD(t *testing.T) {
	input := bytes.NewBufferString("hello3\fworld3\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("\f"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := "┌────────┬────────┐\n│ hello3 │ world3 │\n└────────┴────────┘\n"
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_WithFieldDelimiter_WithFilterIndexes(t *testing.T) {
	input := bytes.NewBufferString("hello|world\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("|"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithFilterIndexes("2,1"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := "┌───────┬───────┐\n│ world │ hello │\n└───────┴───────┘\n"
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_WithFieldDelimiter_WithFilterIndexes_Invalid(t *testing.T) {
	input := bytes.NewBufferString("hello|world\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("|"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithFilterIndexes("2,x"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.Error(t, err)
	assert.ErrorIs(t, err, tablo.ErrInvalidValue)
	assert.Nil(t, tbl)
}

func TestTablo_Tabelize_WithFieldDelimiter_Column_Selection(t *testing.T) {
	input := bytes.NewBufferString("header1.header2\nfoo.bar\n")
	output := new(BytesWriteCloser)

	oldIsNamedPipe := tablo.IsNamedPipe
	oldIsCharDevice := tablo.IsCharDevice
	tablo.IsNamedPipe = func(_ os.FileInfo) bool { return true }
	tablo.IsCharDevice = func(_ os.FileInfo) bool { return false }
	defer func() {
		tablo.IsNamedPipe = oldIsNamedPipe
		tablo.IsCharDevice = oldIsCharDevice
	}()

	tbl, err := tablo.New(
		tablo.WithArgs([]string{"header1"}),
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("."),
		tablo.WithLineDelimiter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := "┌─────────┐\n│ header1 │\n├─────────┤\n│ foo     │\n└─────────┘\n"
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Tabelize_WithFieldDelimiter_Wrong_Column_Selection(t *testing.T) {
	input := bytes.NewBufferString("header1.header2\nfoo.bar\n")
	output := new(BytesWriteCloser)

	oldIsNamedPipe := tablo.IsNamedPipe
	oldIsCharDevice := tablo.IsCharDevice
	tablo.IsNamedPipe = func(_ os.FileInfo) bool { return true }
	tablo.IsCharDevice = func(_ os.FileInfo) bool { return false }
	defer func() {
		tablo.IsNamedPipe = oldIsNamedPipe
		tablo.IsCharDevice = oldIsCharDevice
	}()

	tbl, err := tablo.New(
		tablo.WithArgs([]string{"xxx"}),
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("."),
		tablo.WithLineDelimiter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `┌─────────┬─────────┐
│ header1 │ header2 │
├─────────┼─────────┤
│ foo     │ bar     │
└─────────┴─────────┘
`
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Tabelize_WithFieldDelimiter_Wrong_Column_Selection_WithNoHeaders(t *testing.T) {
	input := bytes.NewBufferString("header1.header2\nfoo.bar\n")
	output := new(BytesWriteCloser)

	oldIsNamedPipe := tablo.IsNamedPipe
	oldIsCharDevice := tablo.IsCharDevice
	tablo.IsNamedPipe = func(_ os.FileInfo) bool { return true }
	tablo.IsCharDevice = func(_ os.FileInfo) bool { return false }
	defer func() {
		tablo.IsNamedPipe = oldIsNamedPipe
		tablo.IsCharDevice = oldIsCharDevice
	}()

	tbl, err := tablo.New(
		tablo.WithArgs([]string{"xxx"}),
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("."),
		tablo.WithLineDelimiter("\n"),
		tablo.WithNoHeaders(true),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := "┌─────┬─────┐\n│ foo │ bar │\n└─────┴─────┘\n"
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Tabelize_FilterIndexesIgnoreArgsForHeaders(t *testing.T) {
	input := bytes.NewBufferString("name|age\nvigo|42\n")
	output := new(BytesWriteCloser)

	oldIsNamedPipe := tablo.IsNamedPipe
	oldIsCharDevice := tablo.IsCharDevice
	tablo.IsNamedPipe = func(_ os.FileInfo) bool { return true }
	tablo.IsCharDevice = func(_ os.FileInfo) bool { return false }
	defer func() {
		tablo.IsNamedPipe = oldIsNamedPipe
		tablo.IsCharDevice = oldIsCharDevice
	}()

	tbl, err := tablo.New(
		tablo.WithArgs([]string{"name"}),
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("|"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithFilterIndexes("2"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := "┌─────┐\n│ age │\n├─────┤\n│ 42  │\n└─────┘\n"
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Tabelize_Docker_Images(t *testing.T) {
	input := bytes.NewBufferString(`REPOSITORY                         TAG       IMAGE ID       CREATED         SIZE
superset-superset-worker-beat      latest    3292fc2e6758   2 weeks ago     958MB
superset-superset-worker           latest    d25cbcc60691   2 weeks ago     958MB
`)
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithLineDelimiter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
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
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_Docker_Images_Filter_By_Header(t *testing.T) {
	input := bytes.NewBufferString(`REPOSITORY                         TAG       IMAGE ID       CREATED         SIZE
superset-superset-worker-beat      latest    3292fc2e6758   2 weeks ago     958MB
superset-superset-worker           latest    d25cbcc60691   2 weeks ago     958MB
`)
	output := new(BytesWriteCloser)

	oldIsNamedPipe := tablo.IsNamedPipe
	oldIsCharDevice := tablo.IsCharDevice
	tablo.IsNamedPipe = func(_ os.FileInfo) bool { return true }
	tablo.IsCharDevice = func(_ os.FileInfo) bool { return false }
	defer func() {
		tablo.IsNamedPipe = oldIsNamedPipe
		tablo.IsCharDevice = oldIsCharDevice
	}()

	tbl, err := tablo.New(
		tablo.WithArgs([]string{"REPOSITORY"}),
		tablo.WithOutputWriter(output),
		tablo.WithLineDelimiter("\n"),
		tablo.WithNoSeparateRows(true),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
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
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Tabelize_Docker_Images_Filter_By_Header_With_No_Border(t *testing.T) {
	input := bytes.NewBufferString(`REPOSITORY                         TAG       IMAGE ID       CREATED         SIZE
superset-superset-worker-beat      latest    3292fc2e6758   2 weeks ago     958MB
superset-superset-worker           latest    d25cbcc60691   2 weeks ago     958MB
`)
	output := new(BytesWriteCloser)

	oldIsNamedPipe := tablo.IsNamedPipe
	oldIsCharDevice := tablo.IsCharDevice
	tablo.IsNamedPipe = func(_ os.FileInfo) bool { return true }
	tablo.IsCharDevice = func(_ os.FileInfo) bool { return false }
	defer func() {
		tablo.IsNamedPipe = oldIsNamedPipe
		tablo.IsCharDevice = oldIsCharDevice
	}()

	tbl, err := tablo.New(
		tablo.WithArgs([]string{"REPOSITORY"}),
		tablo.WithOutputWriter(output),
		tablo.WithLineDelimiter("\n"),
		tablo.WithNoDrawBorder(true),
		tablo.WithNoSeparateRows(true),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `REPOSITORY                    
superset-superset-worker-beat 
superset-superset-worker      
`
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Tabelize_Docker_Images_Filter_By_Header_With_No_Border_With_No_Header(t *testing.T) {
	input := bytes.NewBufferString(`REPOSITORY                         TAG       IMAGE ID       CREATED         SIZE
superset-superset-worker-beat      latest    3292fc2e6758   2 weeks ago     958MB
superset-superset-worker           latest    d25cbcc60691   2 weeks ago     958MB
`)
	output := new(BytesWriteCloser)

	oldIsNamedPipe := tablo.IsNamedPipe
	oldIsCharDevice := tablo.IsCharDevice
	tablo.IsNamedPipe = func(_ os.FileInfo) bool { return true }
	tablo.IsCharDevice = func(_ os.FileInfo) bool { return false }
	defer func() {
		tablo.IsNamedPipe = oldIsNamedPipe
		tablo.IsCharDevice = oldIsCharDevice
	}()

	tbl, err := tablo.New(
		tablo.WithArgs([]string{"REPOSITORY"}),
		tablo.WithOutputWriter(output),
		tablo.WithLineDelimiter("\n"),
		tablo.WithNoDrawBorder(true),
		tablo.WithNoHeaders(true),
		tablo.WithNoSeparateRows(true),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `superset-superset-worker-beat 
superset-superset-worker      
`
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Tabelize_CSV_WithNoHeaders_SkipsDetectedHeaderRow(t *testing.T) {
	input := bytes.NewBufferString(`Username;Identifier;First name;Last name
booker12;9012;Rachel;Booker
grey07;2070;Laura;Grey
`)
	output := new(BytesWriteCloser)

	oldIsNamedPipe := tablo.IsNamedPipe
	oldIsCharDevice := tablo.IsCharDevice
	tablo.IsNamedPipe = func(_ os.FileInfo) bool { return true }
	tablo.IsCharDevice = func(_ os.FileInfo) bool { return false }
	defer func() {
		tablo.IsNamedPipe = oldIsNamedPipe
		tablo.IsCharDevice = oldIsCharDevice
	}()

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter(";"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithNoHeaders(true),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `┌──────────┬──────┬────────┬────────┐
│ booker12 │ 9012 │ Rachel │ Booker │
├──────────┼──────┼────────┼────────┤
│ grey07   │ 2070 │ Laura  │ Grey   │
└──────────┴──────┴────────┴────────┘
`
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Tabelize_CSV_FilterIndexes_WithNoHeaders_SkipsDetectedHeaderRow(t *testing.T) {
	input := bytes.NewBufferString(`Username;Identifier;First name;Last name
booker12;9012;Rachel;Booker
grey07;2070;Laura;Grey
`)
	output := new(BytesWriteCloser)

	oldIsNamedPipe := tablo.IsNamedPipe
	oldIsCharDevice := tablo.IsCharDevice
	tablo.IsNamedPipe = func(_ os.FileInfo) bool { return true }
	tablo.IsCharDevice = func(_ os.FileInfo) bool { return false }
	defer func() {
		tablo.IsNamedPipe = oldIsNamedPipe
		tablo.IsCharDevice = oldIsCharDevice
	}()

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter(";"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithFilterIndexes("1,3"),
		tablo.WithNoHeaders(true),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `┌──────────┬────────┐
│ booker12 │ Rachel │
├──────────┼────────┤
│ grey07   │ Laura  │
└──────────┴────────┘
`
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Tabelize_CSV_WithQuotedUnicodeHeaders_AndNoHeaders_SkipsHeaderRow(t *testing.T) {
	input := bytes.NewBufferString(`ID,"_ADI SOYADI","-T.C. KİMLİK NO",-BÖLÜM
999721,"Filiz Yenilmez",23795235860,Muhasebe
999722,"Ali Veli",12345678901,İnsan Kaynakları
`)
	output := new(BytesWriteCloser)

	oldIsNamedPipe := tablo.IsNamedPipe
	oldIsCharDevice := tablo.IsCharDevice
	tablo.IsNamedPipe = func(_ os.FileInfo) bool { return true }
	tablo.IsCharDevice = func(_ os.FileInfo) bool { return false }
	defer func() {
		tablo.IsNamedPipe = oldIsNamedPipe
		tablo.IsCharDevice = oldIsCharDevice
	}()

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter(","),
		tablo.WithLineDelimiter("\n"),
		tablo.WithNoHeaders(true),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `┌────────┬──────────────────┬─────────────┬──────────────────┐
│ 999721 │ "Filiz Yenilmez" │ 23795235860 │ Muhasebe         │
├────────┼──────────────────┼─────────────┼──────────────────┤
│ 999722 │ "Ali Veli"       │ 12345678901 │ İnsan Kaynakları │
└────────┴──────────────────┴─────────────┴──────────────────┘
`
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Tabelize_CSV_AutoDetectsDelimiter_WithNoHeaders(t *testing.T) {
	input := bytes.NewBufferString(`ID,"_ADI SOYADI","-T.C. KİMLİK NO",-BÖLÜM
999721,"Filiz Yenilmez",23795235860,Muhasebe
999722,"Ali Veli",12345678901,İnsan Kaynakları
`)
	output := new(BytesWriteCloser)

	oldIsNamedPipe := tablo.IsNamedPipe
	oldIsCharDevice := tablo.IsCharDevice
	tablo.IsNamedPipe = func(_ os.FileInfo) bool { return true }
	tablo.IsCharDevice = func(_ os.FileInfo) bool { return false }
	defer func() {
		tablo.IsNamedPipe = oldIsNamedPipe
		tablo.IsCharDevice = oldIsCharDevice
	}()

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithLineDelimiter("\n"),
		tablo.WithNoHeaders(true),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `┌────────┬──────────────────┬─────────────┬──────────────────┐
│ 999721 │ "Filiz Yenilmez" │ 23795235860 │ Muhasebe         │
├────────┼──────────────────┼─────────────┼──────────────────┤
│ 999722 │ "Ali Veli"       │ 12345678901 │ İnsan Kaynakları │
└────────┴──────────────────┴─────────────┴──────────────────┘
`
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Run_Returns_Error(t *testing.T) {
	os.Args = []string{"tablo", "-l", ""}
	resetFlags()

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := tablo.Run()
	assert.Error(t, err)
	_ = w.Close()
	os.Stdout = oldStdout
}

func TestTablo_Run_Read_Input_From_File(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	content := `SCENE       SCENER            GROUP
C64         vigo              Bronx^zOMBiE BoYS^tRSI
AMIGA       turbo             Bronx^zOMBiE BoYS
AMIGA       move              Bronx^zOMBiE BoYS
C64         street tuff       tRSI`

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	_ = tmpFile.Close()

	os.Args = []string{"tablo", "-n", tmpFile.Name()}
	resetFlags()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = tablo.Run()
	assert.NoError(t, err)
	_ = w.Close()
	os.Stdout = oldStdout

	output := new(BytesWriteCloser)
	_, _ = output.ReadFrom(r)

	expectedOutput := `┌───────┬─────────────┬────────────────────────┐
│ SCENE │ SCENER      │ GROUP                  │
│ C64   │ vigo        │ Bronx^zOMBiE BoYS^tRSI │
│ AMIGA │ turbo       │ Bronx^zOMBiE BoYS      │
│ AMIGA │ move        │ Bronx^zOMBiE BoYS      │
│ C64   │ street tuff │ tRSI                   │
└───────┴─────────────┴────────────────────────┘
`
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Run_Read_Input_From_File_Filter_Header(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	content := `# comment
SCENE       SCENER            GROUP
C64         vigo              Bronx^zOMBiE BoYS^tRSI
AMIGA       turbo             Bronx^zOMBiE BoYS
AMIGA       move              Bronx^zOMBiE BoYS
C64         street tuff       tRSI`

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	_ = tmpFile.Close()

	os.Args = []string{"tablo", "-n", tmpFile.Name(), "SCENE"}
	resetFlags()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = tablo.Run()
	assert.NoError(t, err)
	_ = w.Close()
	os.Stdout = oldStdout

	output := new(BytesWriteCloser)
	_, _ = output.ReadFrom(r)

	expectedOutput := `┌───────┐
│ SCENE │
├───────┤
│ C64   │
│ AMIGA │
│ AMIGA │
│ C64   │
└───────┘
`
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Run_Read_Input_From_File_TAB_as_LineDelimiter(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	content := "hello\tworld"

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	_ = tmpFile.Close()

	os.Args = []string{"tablo", "-n", "-l", "\t", tmpFile.Name()}
	resetFlags()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = tablo.Run()
	assert.NoError(t, err)
	_ = w.Close()
	os.Stdout = oldStdout

	output := new(BytesWriteCloser)
	_, _ = output.ReadFrom(r)

	expectedOutput := `┌───────┐
│ hello │
│ world │
└───────┘
`
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Run_Read_Input_From_File_R_as_LineDelimiter(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	content := "hello\rworld"

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	_ = tmpFile.Close()

	os.Args = []string{"tablo", "-n", "-l", "\r", tmpFile.Name()}
	resetFlags()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = tablo.Run()
	assert.NoError(t, err)
	_ = w.Close()
	os.Stdout = oldStdout

	output := new(BytesWriteCloser)
	_, _ = output.ReadFrom(r)

	expectedOutput := `┌───────┐
│ hello │
│ world │
└───────┘
`
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Run_Read_Input_From_Non_Existing_File(t *testing.T) {
	os.Args = []string{"tablo", "f4|<3"}
	resetFlags()

	oldStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	err := tablo.Run()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, fs.ErrNotExist))
	_ = w.Close()
	os.Stderr = oldStderr
}

func TestRun_ReadFromFile_SaveToFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_output.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := `this is vigo
`
	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	_ = tmpFile.Close()

	outputFile, err := os.CreateTemp("", "output.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(outputFile.Name()) }()

	os.Args = []string{"tablo", "-o", outputFile.Name(), tmpFile.Name()}
	resetFlags()

	err = tablo.Run()
	assert.NoError(t, err)

	result, err := os.ReadFile(outputFile.Name())
	assert.NoError(t, err)

	expectedOutput := `┌──────────────┐
│ this is vigo │
└──────────────┘
`
	assert.Equal(t, expectedOutput, string(result))
}

func TestTablo_Tabelize_JSONOutput_WithHeaders(t *testing.T) {
	input := bytes.NewBufferString("name|age\nvigo|42\njohn|7\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("|"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithJSONOutput(true),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `[
  {
    "name": "vigo",
    "age": "42"
  },
  {
    "name": "john",
    "age": "7"
  }
]
`
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_JSONOutput_WithoutHeaders(t *testing.T) {
	input := bytes.NewBufferString("hello|world\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter("|"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithFilterIndexes("2,1"),
		tablo.WithJSONOutput(true),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `[
  [
    "world",
    "hello"
  ]
]
`
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestRun_ReadFromFile_SaveToJSONFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_input.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString("name|age\nvigo|42\n")
	assert.NoError(t, err)
	_ = tmpFile.Close()

	outputFile, err := os.CreateTemp("", "output.json")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(outputFile.Name()) }()

	os.Args = []string{"tablo", "-j", "-f", "|", "-o", outputFile.Name(), tmpFile.Name()}
	resetFlags()

	err = tablo.Run()
	assert.NoError(t, err)

	result, err := os.ReadFile(outputFile.Name())
	assert.NoError(t, err)

	expectedOutput := `[
  {
    "name": "vigo",
    "age": "42"
  }
]
`
	assert.Equal(t, expectedOutput, string(result))
}

func TestTablo_Tabelize_JSONOutput_WithoutHeaderRow_KeepsFirstRow(t *testing.T) {
	input := bytes.NewBufferString("nobody:*:-2:-2:Unprivileged User:/var/empty:/usr/bin/false\nroot:*:0:0:System Administrator:/var/root:/bin/sh\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter(":"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithJSONOutput(true),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := `[
  [
    "nobody",
    "*",
    "-2",
    "-2",
    "Unprivileged User",
    "/var/empty",
    "/usr/bin/false"
  ],
  [
    "root",
    "*",
    "0",
    "0",
    "System Administrator",
    "/var/root",
    "/bin/sh"
  ]
]
`
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_JSONOutput_EmptyInput_ReturnsArray(t *testing.T) {
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithLineDelimiter("\n"),
		tablo.WithJSONOutput(true),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return "", nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	assert.Equal(t, "[]\n", string(output.nonStdinValue()))
}

func TestTablo_Tabelize_FieldDelimiter_Space_ExactSplit(t *testing.T) {
	input := bytes.NewBufferString("foo bar\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter(" "),
		tablo.WithLineDelimiter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := "┌─────┬─────┐\n│ foo │ bar │\n└─────┴─────┘\n"
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_FieldDelimiter_Colon_PreservesEmptyFields(t *testing.T) {
	input := bytes.NewBufferString("a:b::c\n")
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithFieldDelimiter(":"),
		tablo.WithLineDelimiter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
		}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, tbl)

	err = tbl.Tabelize()
	assert.NoError(t, err)

	expectedOutput := "┌───┬───┬──┬───┐\n│ a │ b │  │ c │\n└───┴───┴──┴───┘\n"
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Tabelize_DefaultFieldDelimiter_Uses_SmartMode_For_Aligned_Output(t *testing.T) {
	input := bytes.NewBufferString(`REPOSITORY                         TAG       IMAGE ID       CREATED         SIZE
superset-superset-worker-beat      latest    3292fc2e6758   2 weeks ago     958MB
superset-superset-worker           latest    d25cbcc60691   2 weeks ago     958MB
`)
	output := new(BytesWriteCloser)

	tbl, err := tablo.New(
		tablo.WithOutputWriter(output),
		tablo.WithLineDelimiter("\n"),
		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
			return input.String(), nil
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
	assert.Equal(t, expectedOutput, string(output.nonStdinValue()))
}

func TestTablo_Run_FieldDelimiter_Space_ExactSplit_FromCLI(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString("foo bar")
	assert.NoError(t, err)
	_ = tmpFile.Close()

	os.Args = []string{"tablo", "-n", "-f", " ", tmpFile.Name()}
	resetFlags()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = tablo.Run()
	assert.NoError(t, err)
	_ = w.Close()
	os.Stdout = oldStdout

	output := new(BytesWriteCloser)
	_, _ = output.ReadFrom(r)

	expectedOutput := `┌─────┬─────┐
│ foo │ bar │
└─────┴─────┘
`
	assert.Equal(t, expectedOutput, output.String())
}

func TestTablo_Run_FieldDelimiter_Colon_PreservesEmptyFields_FromCLI(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test.txt")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString("a:b::c")
	assert.NoError(t, err)
	_ = tmpFile.Close()

	os.Args = []string{"tablo", "-n", "-f", ":", tmpFile.Name()}
	resetFlags()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = tablo.Run()
	assert.NoError(t, err)
	_ = w.Close()
	os.Stdout = oldStdout

	output := new(BytesWriteCloser)
	_, _ = output.ReadFrom(r)

	expectedOutput := `┌───┬───┬──┬───┐
│ a │ b │  │ c │
└───┴───┴──┴───┘
`
	assert.Equal(t, expectedOutput, output.String())
}

func TestRun_ShowUsage_WithHelpFlag(t *testing.T) {
	os.Args = []string{"tablo", "-h"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	var buf bytes.Buffer
	flag.CommandLine.SetOutput(&buf)

	err := tablo.Run()

	assert.ErrorIs(t, err, flag.ErrHelp)
	assert.Contains(t, buf.String(), "usage:")
	assert.Contains(t, buf.String(), "tablo")
	assert.Contains(t, buf.String(), "-bash-completion")
	assert.Contains(t, buf.String(), "--bash-completion")
}

func TestRun_PrintBashCompletion(t *testing.T) {
	os.Args = []string{"tablo", "--bash-completion"}
	resetFlags()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	err = tablo.Run()
	assert.NoError(t, err)
	_ = w.Close()

	output := new(BytesWriteCloser)
	_, _ = output.ReadFrom(r)

	assert.Contains(t, output.String(), "complete -F _tablo_completion -- 'tablo'")
	assert.Contains(t, output.String(), "--complete")
}

func TestRun_PrintBashCompletion_AfterOtherFlags(t *testing.T) {
	os.Args = []string{"tablo", "-n", "--bash-completion"}
	resetFlags()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	err = tablo.Run()
	assert.NoError(t, err)
	_ = w.Close()

	output := new(BytesWriteCloser)
	_, _ = output.ReadFrom(r)

	assert.Contains(t, output.String(), "complete -F _tablo_completion -- 'tablo'")
}

func TestRun_PrintBashCompletion_ShortLongFlag(t *testing.T) {
	os.Args = []string{"tablo", "-bash-completion"}
	resetFlags()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	err = tablo.Run()
	assert.NoError(t, err)
	_ = w.Close()

	output := new(BytesWriteCloser)
	_, _ = output.ReadFrom(r)

	assert.Contains(t, output.String(), "complete -F _tablo_completion -- 'tablo'")
}
