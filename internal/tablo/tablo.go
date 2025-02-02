/*
Package tablo implements tablized out for given file or piped data.
*/
package tablo

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

var _ Tablizer = (*Tablo)(nil) // compile time proof

const (
	breakTextForUnix    = "press ENTER and CTRL+D to finish text entry"
	breakTextForWindows = "press CTRL+Z and ENTER to finish text entry"

	helpOutput             = "where to send output"
	helpLineDelimeterChar  = "line delimiter char to split the input"
	helpFieldDelimeterChar = "field delimiter char to split the line input"
	helpSeparateRows       = "draw separation line under rows"

	defaultOutput         = "stdout"
	defaultLineDelimeter  = '\n'
	defaultFieldDelimeter = ' '
)

// sentinel errors.
var (
	ErrValueRequired = errors.New("value required")

	spaceSplitter = regexp.MustCompile(`\s{2,}`)
)

// Tablizer defines main functionality.
type Tablizer interface {
	Tabelize() error
}

// ReadInputFunc is a function type.
type ReadInputFunc func(io.Reader) (string, error)

// ParseArgsFunc ia a function type.
type ParseArgsFunc func([]string) ([]string, string)

func readInput(input io.Reader) (string, error) {
	r := bufio.NewReader(input)
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	if len(b) == 0 {
		return "", nil
	}

	return string(b[:len(b)-1]), nil
}

func parseArgs(args []string) ([]string, string) {
	var possibleFile string

	if len(args) == 0 {
		return nil, possibleFile
	}

	lastArg := args[len(args)-1]

	fileInfo, err := os.Stat(lastArg)
	if err == nil && !fileInfo.IsDir() {
		possibleFile = lastArg
	}

	return args, possibleFile
}

func stringSliceToRow(fields []string) table.Row {
	row := make(table.Row, len(fields))
	for i, v := range fields {
		row[i] = v
	}
	return row
}

// Run runs the command.
func Run() error {
	output := flag.String("output", defaultOutput, helpOutput)
	flag.StringVar(output, "o", defaultOutput, helpOutput+" (short)")

	version := flag.Bool("version", false, "display version information")

	lineDelimiterChar := flag.String("line-delimeter-char", string(defaultLineDelimeter), helpLineDelimeterChar)
	flag.StringVar(lineDelimiterChar, "l", string(defaultLineDelimeter), helpLineDelimeterChar+" (short)")

	fieldDelimiterChar := flag.String("field-delimeter-char", string(defaultFieldDelimeter), helpFieldDelimeterChar)
	flag.StringVar(fieldDelimiterChar, "f", string(defaultFieldDelimeter), helpFieldDelimeterChar+" (short)")

	separateRows := flag.Bool("no-separate-rows", false, helpSeparateRows)
	flag.BoolVar(separateRows, "nsr", false, helpSeparateRows+" (short)")

	flag.Parse()

	tbl, err := New(
		WithArgs(flag.Args()),
		WithOutput(*output),
		WithDisplayVersion(*version),
		WithParseArgsFunc(parseArgs),
		WithReadInputFunc(readInput),
		WithLineDelimeter(*lineDelimiterChar),
		WithFieldDelimeter(*fieldDelimiterChar),
		WithSeparateRows(*separateRows),
	)
	if err != nil {
		return err
	}
	defer func() { _ = tbl.Output.Close() }()

	if err = tbl.Tabelize(); err != nil {
		return err
	}

	if *output != "stdout" {
		fmt.Fprintf(flag.CommandLine.Output(), "result saved to: %s\n", *output)
	}

	return nil
}

// Tablo holds the required params.
type Tablo struct {
	Version        string
	Output         io.WriteCloser
	ReadInputFunc  ReadInputFunc
	ParseArgsFunc  ParseArgsFunc
	Args           []string
	LineDelimeter  rune
	FieldDelimeter rune
	DisplayVersion bool
	SeparateRows   bool
}

func (t *Tablo) setDefaults() {
	if t.Output == nil {
		t.Output = os.Stdout
	}
	if t.ReadInputFunc == nil {
		t.ReadInputFunc = readInput
	}
	if t.ParseArgsFunc == nil {
		t.ParseArgsFunc = parseArgs
	}
	t.Version = Version
}

// Tabelize generates tablized output.
func (t *Tablo) Tabelize() error {
	if t.DisplayVersion {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n", t.Version)
		return nil
	}

	args, fileArg := t.ParseArgsFunc(t.Args)
	t.Args = args

	readFrom := os.Stdin
	if fileArg != "" {
		file, err := os.Open(filepath.Clean(fileArg))
		if err != nil {
			return err
		}

		readFrom = file
		defer func() { _ = file.Close() }()
	}

	stat, err := readFrom.Stat()
	if err != nil {
		return err
	}

	if stat.Mode()&os.ModeCharDevice != 0 {
		if runtime.GOOS == "windows" {
			fmt.Fprintln(t.Output, breakTextForWindows)
		} else {
			fmt.Fprintln(t.Output, breakTextForUnix)
		}
	}

	input, err := t.ReadInputFunc(readFrom)
	if err != nil {
		return err
	}

	lines := strings.FieldsFunc(input, func(r rune) bool {
		return r == t.LineDelimeter
	})

	tw := table.NewWriter()
	tw.SetOutputMirror(t.Output)
	tw.SetStyle(table.StyleLight)
	tw.Style().Options.SeparateRows = !t.SeparateRows

	var headers []string
	var columnIndices []int

	for i, line := range lines {
		if len(t.Args) > 0 && i == 0 {
			if t.FieldDelimeter == ' ' {
				headers = spaceSplitter.Split(line, -1)
			} else {
				headers = strings.Split(line, string(t.FieldDelimeter))
			}

			for _, arg := range t.Args {
				for idx, header := range headers {
					if strings.EqualFold(header, arg) {
						columnIndices = append(columnIndices, idx)
						break
					}
				}
			}

			if len(columnIndices) > 0 {
				var selectedHeaders []string
				for _, idx := range columnIndices {
					if idx < len(headers) {
						selectedHeaders = append(selectedHeaders, headers[idx])
					}
				}
				tw.AppendHeader(stringSliceToRow(selectedHeaders))
			} else {
				tw.AppendHeader(stringSliceToRow(headers))
			}

			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		var fields []string
		if t.FieldDelimeter == ' ' {
			fields = spaceSplitter.Split(line, -1)
		} else {
			fields = strings.Split(line, string(t.FieldDelimeter))
		}

		if len(columnIndices) > 0 {
			var selectedFields []string
			for _, idx := range columnIndices {
				if idx < len(fields) {
					selectedFields = append(selectedFields, fields[idx])
				} else {
					selectedFields = append(selectedFields, "x")
				}
			}
			tw.AppendRow(stringSliceToRow(selectedFields))
		} else {
			tw.AppendRow(stringSliceToRow(fields))
		}
	}

	tw.Render()

	return nil
}

// Option represents option function type.
type Option func(*Tablo) error

// WithArgs sets the args.
func WithArgs(args []string) Option {
	return func(t *Tablo) error {
		t.Args = args

		return nil
	}
}

// WithOutput sets the output writer.
func WithOutput(output string) Option {
	return func(t *Tablo) error {
		t.Output = os.Stdout
		if output != "stdout" {
			f, err := os.Create(filepath.Clean(output))
			if err != nil {
				return err
			}
			t.Output = f
		}

		return nil
	}
}

// WithOutputWriter sets the output write for test usage.
func WithOutputWriter(wr io.WriteCloser) Option {
	return func(t *Tablo) error {
		t.Output = wr

		return nil
	}
}

// WithDisplayVersion sets the display version information or not.
func WithDisplayVersion(display bool) Option {
	return func(t *Tablo) error {
		t.DisplayVersion = display

		return nil
	}
}

// WithReadInputFunc sets the read input function.
func WithReadInputFunc(fn ReadInputFunc) Option {
	return func(t *Tablo) error {
		if fn == nil {
			return fmt.Errorf("%w, read input function is nil", ErrValueRequired)
		}
		t.ReadInputFunc = fn

		return nil
	}
}

// WithParseArgsFunc sets the parse args function.
func WithParseArgsFunc(fn ParseArgsFunc) Option {
	return func(t *Tablo) error {
		if fn == nil {
			return fmt.Errorf("%w, parse args function is nil", ErrValueRequired)
		}
		t.ParseArgsFunc = fn

		return nil
	}
}

// WithLineDelimeter sets the line delimeter.
func WithLineDelimeter(s string) Option {
	return func(t *Tablo) error {
		if s == "" {
			return fmt.Errorf("%w, line delimeter is empty string", ErrValueRequired)
		}

		var delimeter rune

		switch s {
		case `\n`:
			delimeter = defaultLineDelimeter
		case `\r`:
			delimeter = '\r'
		case `\t`:
			delimeter = '\t'
		default:
			delimeter = []rune(s)[0]
		}

		t.LineDelimeter = delimeter

		return nil
	}
}

// WithFieldDelimeter sets the field delimeter.
func WithFieldDelimeter(s string) Option {
	return func(t *Tablo) error {
		if s == "" {
			t.FieldDelimeter = 0

			return nil
		}

		var delimeter rune

		switch s {
		case `\f`:
			delimeter = '\f'
		case `\v`:
			delimeter = '\v'
		case `\t`:
			delimeter = '\t'
		default:
			delimeter = []rune(s)[0]
		}

		t.FieldDelimeter = delimeter

		return nil
	}
}

// WithSeparateRows enables/disables the separation row line.
func WithSeparateRows(sep bool) Option {
	return func(t *Tablo) error {
		t.SeparateRows = sep

		return nil
	}
}

// New instantiates new Tablo instance.
func New(options ...Option) (*Tablo, error) {
	tbl := new(Tablo)

	for _, option := range options {
		if err := option(tbl); err != nil {
			return nil, err
		}
	}

	tbl.setDefaults()

	return tbl, nil
}
