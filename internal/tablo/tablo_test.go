package tablo_test

import (
	"bytes"
	"flag"
	"fmt"
	"io"
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

func TestTablo_New_WithParseArgsFunc_Nil(t *testing.T) {
	tbl, err := tablo.New(
		tablo.WithParseArgsFunc(nil),
	)

	assert.Nil(t, tbl)
	assert.Error(t, err)
	assert.ErrorIs(t, err, tablo.ErrValueRequired)
}

func TestTablo_New_WithParseArgsFunc(t *testing.T) {
	tbl, err := tablo.New(
		tablo.WithParseArgsFunc(func(_ []string) ([]string, string) {
			return nil, ""
		}),
	)

	assert.NotNil(t, tbl)
	assert.NoError(t, err)
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

	expectedOutput := "┌─────────────┐\n│ hello world │\n└─────────────┘\n"

	fmt.Println("expectedOutput", expectedOutput)
	fmt.Println("output", string(output.chomp()))
}

// func TestMain_Run_PipedInput(t *testing.T) {
// 	input := bytes.NewBufferString("hello world\n")
// 	output := new(BytesWriteCloser)
//
// 	tbl, err := tablo.New(
// 		tablo.WithOutputWriter(output),
// 		tablo.WithFieldDelimiter(" "),
// 		tablo.WithLineDelimiter("\n"),
// 		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
// 			return input.String(), nil
// 		}),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, tbl)
//
// 	err = tbl.Tabelize()
// 	assert.NoError(t, err)
//
// 	// Expected table output
// 	expectedOutput := `┌─────────────┐
// │ hello world │
// └─────────────┘
// `
// 	lines := strings.Split(output.String(), "\n")
// 	assert.Equal(t, expectedOutput, strings.Join(lines[1:], "\n"))
// }
//
// func TestMain_Run_PipedInputWithFilterIndexes(t *testing.T) {
// 	input := bytes.NewBufferString("hello|world")
// 	output := new(BytesWriteCloser)
//
// 	tbl, err := tablo.New(
// 		tablo.WithOutputWriter(output),
// 		tablo.WithLineDelimiter("\n"),
// 		tablo.WithFieldDelimiter("|"),
// 		tablo.WithFilterIndexes("1"),
// 		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
// 			return input.String(), nil
// 		}),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, tbl)
//
// 	err = tbl.Tabelize()
// 	assert.NoError(t, err)
//
// 	// Expected table output
// 	expectedOutput := `┌───────┐
// │ hello │
// └───────┘
// `
// 	lines := strings.Split(output.String(), "\n")
// 	assert.Equal(t, expectedOutput, strings.Join(lines[1:], "\n"))
// }
//
// func TestMain_Run_PipedInputHeaderWithCustomDelimiter(t *testing.T) {
// 	input := bytes.NewBufferString("header1.......header2\nfoo...........bar\n")
// 	output := new(BytesWriteCloser)
//
// 	tbl, err := tablo.New(
// 		tablo.WithArgs([]string{"header1"}),
// 		tablo.WithOutputWriter(output),
// 		tablo.WithLineDelimiter("\n"),
// 		tablo.WithFieldDelimiter("."),
// 		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
// 			return input.String(), nil
// 		}),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, tbl)
//
// 	err = tbl.Tabelize()
// 	assert.NoError(t, err)
//
// 	expectedOutput := `┌─────────┐
// │ header1 │
// ├─────────┤
// │ foo     │
// └─────────┘
// `
// 	lines := strings.Split(output.String(), "\n")
// 	assert.Equal(t, expectedOutput, strings.Join(lines[1:], "\n"))
// }
//
// func TestMain_Run_PipedInputHeaderWithCustomDelimiter_WrongHeader(t *testing.T) {
// 	input := bytes.NewBufferString("header1.......header2\nfoo...........bar\n")
// 	output := new(BytesWriteCloser)
//
// 	tbl, err := tablo.New(
// 		tablo.WithArgs([]string{"headerx"}),
// 		tablo.WithOutputWriter(output),
// 		tablo.WithLineDelimiter("\n"),
// 		tablo.WithFieldDelimiter("."),
// 		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
// 			return input.String(), nil
// 		}),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, tbl)
//
// 	err = tbl.Tabelize()
// 	assert.NoError(t, err)
//
// 	expectedOutput := `┌─────────┬─────────┐
// │ header1 │ header2 │
// ├─────────┼─────────┤
// │ foo     │ bar     │
// └─────────┴─────────┘
// `
// 	lines := strings.Split(output.String(), "\n")
// 	assert.Equal(t, expectedOutput, strings.Join(lines[1:], "\n"))
// }
//
// func TestMain_Run_PipedInput_with_FieldDelimiter(t *testing.T) {
// 	input := bytes.NewBufferString("hello:world\n")
// 	output := new(BytesWriteCloser)
//
// 	tbl, err := tablo.New(
// 		tablo.WithOutputWriter(output),
// 		tablo.WithLineDelimiter("\n"),
// 		tablo.WithFieldDelimiter(":"),
// 		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
// 			return input.String(), nil
// 		}),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, tbl)
//
// 	err = tbl.Tabelize()
// 	assert.NoError(t, err)
//
// 	expectedOutput := `┌───────┬───────┐
// │ hello │ world │
// └───────┴───────┘
// `
// 	lines := strings.Split(output.String(), "\n")
// 	assert.Equal(t, expectedOutput, strings.Join(lines[1:], "\n"))
// }
//
// func TestMain_Run_PipedInput_docker_images(t *testing.T) {
// 	input := bytes.NewBufferString(`REPOSITORY                         TAG       IMAGE ID       CREATED         SIZE
// superset-superset-worker-beat      latest    3292fc2e6758   2 weeks ago     958MB
// superset-superset-worker           latest    d25cbcc60691   2 weeks ago     958MB
// `)
// 	output := new(BytesWriteCloser)
//
// 	tbl, err := tablo.New(
// 		tablo.WithOutputWriter(output),
// 		tablo.WithFieldDelimiter(" "),
// 		tablo.WithLineDelimiter("\n"),
// 		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
// 			return input.String(), nil
// 		}),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, tbl)
//
// 	err = tbl.Tabelize()
// 	assert.NoError(t, err)
//
// 	expectedOutput := `┌───────────────────────────────┬────────┬──────────────┬─────────────┬───────┐
// │ REPOSITORY                    │ TAG    │ IMAGE ID     │ CREATED     │ SIZE  │
// ├───────────────────────────────┼────────┼──────────────┼─────────────┼───────┤
// │ superset-superset-worker-beat │ latest │ 3292fc2e6758 │ 2 weeks ago │ 958MB │
// ├───────────────────────────────┼────────┼──────────────┼─────────────┼───────┤
// │ superset-superset-worker      │ latest │ d25cbcc60691 │ 2 weeks ago │ 958MB │
// └───────────────────────────────┴────────┴──────────────┴─────────────┴───────┘
// `
// 	lines := strings.Split(output.String(), "\n")
// 	assert.Equal(t, expectedOutput, strings.Join(lines[1:], "\n"))
// }
//
// func TestMain_Run_PipedInput_docker_images_with_args(t *testing.T) {
// 	input := bytes.NewBufferString(`REPOSITORY                         TAG       IMAGE ID       CREATED         SIZE
// superset-superset-worker-beat      latest    3292fc2e6758   2 weeks ago     958MB
// superset-superset-worker           latest    d25cbcc60691   2 weeks ago     958MB
// `)
// 	output := new(BytesWriteCloser)
//
// 	tbl, err := tablo.New(
// 		tablo.WithOutputWriter(output),
// 		tablo.WithFieldDelimiter(" "),
// 		tablo.WithLineDelimiter("\n"),
// 		tablo.WithArgs([]string{"REPOSITORY"}),
// 		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
// 			return input.String(), nil
// 		}),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, tbl)
//
// 	err = tbl.Tabelize()
// 	assert.NoError(t, err)
//
// 	expectedOutput := `┌───────────────────────────────┐
// │ REPOSITORY                    │
// ├───────────────────────────────┤
// │ superset-superset-worker-beat │
// ├───────────────────────────────┤
// │ superset-superset-worker      │
// └───────────────────────────────┘
// `
// 	lines := strings.Split(output.String(), "\n")
// 	assert.Equal(t, expectedOutput, strings.Join(lines[1:], "\n"))
// }
//
// func TestMain_Run_PipedInput_docker_images_with_no_separation(t *testing.T) {
// 	input := bytes.NewBufferString(`REPOSITORY                         TAG       IMAGE ID       CREATED         SIZE
// superset-superset-worker-beat      latest    3292fc2e6758   2 weeks ago     958MB
// superset-superset-worker           latest    d25cbcc60691   2 weeks ago     958MB
// `)
// 	output := new(BytesWriteCloser)
//
// 	tbl, err := tablo.New(
// 		tablo.WithOutputWriter(output),
// 		tablo.WithFieldDelimiter(" "),
// 		tablo.WithLineDelimiter("\n"),
// 		tablo.WithSeparateRows(true),
// 		tablo.WithArgs([]string{"REPOSITORY"}),
// 		tablo.WithReadInputFunc(func(_ io.Reader) (string, error) {
// 			return input.String(), nil
// 		}),
// 	)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, tbl)
//
// 	err = tbl.Tabelize()
// 	assert.NoError(t, err)
//
// 	expectedOutput := `┌───────────────────────────────┐
// │ REPOSITORY                    │
// ├───────────────────────────────┤
// │ superset-superset-worker-beat │
// │ superset-superset-worker      │
// └───────────────────────────────┘
// `
// 	lines := strings.Split(output.String(), "\n")
// 	assert.Equal(t, expectedOutput, strings.Join(lines[1:], "\n"))
// }
//
//
// func TestRun_ReadFromFile(t *testing.T) {
// 	tmpFile, err := os.CreateTemp("", "test.txt")
// 	assert.NoError(t, err)
// 	defer func() { _ = os.Remove(tmpFile.Name()) }()
//
// 	content := `this is vigo
// `
// 	_, err = tmpFile.WriteString(content)
// 	assert.NoError(t, err)
// 	_ = tmpFile.Close()
//
// 	err = tablo.Run()
// 	assert.NoError(t, err)
//
// 	os.Args = []string{"tablo", tmpFile.Name()}
// 	resetFlags()
//
// 	oldStdout := os.Stdout
// 	r, w, _ := os.Pipe()
// 	os.Stdout = w
//
// 	err = tablo.Run()
// 	assert.NoError(t, err)
// 	_ = w.Close()
// 	os.Stdout = oldStdout
//
// 	output := new(BytesWriteCloser)
// 	_, _ = output.ReadFrom(r)
//
// 	expectedOutput := `┌──────────────┐
// │ this is vigo │
// └──────────────┘
// `
//
// 	assert.Equal(t, expectedOutput, output.String())
// }
//
// // func TestRun_ReadFrom_NonExistingFile(t *testing.T) {
// // 	os.Args = []string{"tablo", "fake-file"}
// // 	resetFlags()
// //
// // 	err := tablo.Run()
// // 	assert.Error(t, err)
// // }
//
// func TestRun_ReadFromFile_SaveToFile(t *testing.T) {
// 	tmpFile, err := os.CreateTemp("", "test_output.txt")
// 	assert.NoError(t, err)
// 	defer func() { _ = os.Remove(tmpFile.Name()) }()
//
// 	content := `this is vigo
// `
// 	_, err = tmpFile.WriteString(content)
// 	assert.NoError(t, err)
// 	_ = tmpFile.Close()
//
// 	outputFile, err := os.CreateTemp("", "output.txt")
// 	assert.NoError(t, err)
// 	defer func() { _ = os.Remove(outputFile.Name()) }()
//
// 	os.Args = []string{"tablo", "-o", outputFile.Name(), tmpFile.Name()}
// 	resetFlags()
//
// 	err = tablo.Run()
// 	assert.NoError(t, err)
//
// 	result, err := os.ReadFile(outputFile.Name())
// 	assert.NoError(t, err)
//
// 	expectedOutput := `┌──────────────┐
// │ this is vigo │
// └──────────────┘
// `
// 	assert.Equal(t, expectedOutput, string(result))
// }
