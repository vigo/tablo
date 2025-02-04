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
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var _ Tablizer = (*Tablo)(nil) // compile time proof

const (
	breakTextForUnix    = "press ENTER and CTRL+D to finish text entry"
	breakTextForWindows = "press CTRL+Z and ENTER to finish text entry"

	helpOutput             = "where to send output"
	helpLineDelimiterChar  = "line delimiter char to split the input"
	helpFieldDelimiterChar = "field delimiter char to split the line input"
	helpNoSeparateRows     = "do not draw separation line under rows"
	helpFilterIndexes      = "filter columns by index"

	defaultOutput         = "stdout"
	defaultFieldDelimiter = ' '
	defaultLineDelimiter  = '\n'
	defaultSpaceAmount    = 2
)

// sentinel errors.
var (
	ErrValueRequired = errors.New("value required")
	ErrInvalidValue  = errors.New("invalid value")
	ErrInvalidFile   = errors.New("invalid file")
)

// Tablizer defines main functionality.
type Tablizer interface {
	Tabelize() error
}

func spaceSplitter(spaceAmount int) *regexp.Regexp {
	return regexp.MustCompile(`\s{` + strconv.Itoa(spaceAmount) + `,}`)
}

func charSplitter(char rune, minRepeat int) *regexp.Regexp {
	pattern := regexp.QuoteMeta(string(char)) + "{" + strconv.Itoa(minRepeat) + ",}"
	return regexp.MustCompile(pattern)
}

func maxConsecutiveRepeats(s string, delimiter rune) int {
	pattern := regexp.MustCompile(regexp.QuoteMeta(string(delimiter)) + "+")
	matches := pattern.FindAllString(s, -1)

	maxCount := 1
	for _, match := range matches {
		if len(match) > maxCount {
			maxCount = len(match)
		}
	}
	return maxCount
}

// ReadInputFunc is a function type.
type ReadInputFunc func(io.Reader) (string, error)

func isFile(filename string) error {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return err
	}

	fileMode := fileInfo.Mode()
	if fileMode.IsRegular() {
		return nil
	}

	return ErrInvalidFile
}

func readInput(input io.Reader) (string, error) {
	r := bufio.NewReader(input)
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	if len(b) == 0 {
		return "", nil
	}

	str := strings.TrimSuffix(string(b), "\n")

	return str, nil
}

func stringSliceToRow(fields []string) table.Row {
	row := make(table.Row, len(fields))
	for i, v := range fields {
		row[i] = v
	}
	return row
}

// Tablo holds the required params.
type Tablo struct {
	Version        string
	Output         io.WriteCloser
	ReadInputFunc  ReadInputFunc
	Args           []string
	FilterIndexes  []int
	LineDelimiter  rune
	FieldDelimiter rune
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
	t.Version = Version
}

func getPipeCharDeviceMode() (fs.FileMode, error) {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return fs.ModeIrregular, err
	}
	return fileInfo.Mode() & os.ModeCharDevice, nil
}

func (t *Tablo) parseArgs() (string, error) {
	pipeMode, pipeErr := getPipeCharDeviceMode()
	if pipeErr != nil {
		return "", pipeErr
	}

	fmt.Println("t.Args", t.Args, len(t.Args))
	fmt.Println("pipeMode", pipeMode)
	fmt.Printf("pipeMode: %d\n", pipeMode)

	if len(t.Args) == 0 || pipeMode == 0 {
		return "", nil
	}

	fmt.Println("t.Args", t.Args)

	fileArg := t.Args[0]
	if err := isFile(fileArg); err != nil {
		return "", err
	}

	t.Args = t.Args[1:]

	return fileArg, nil
}

func (t *Tablo) getReadFrom() (*os.File, error) {
	fileArg, err := t.parseArgs()
	if err != nil {
		return nil, err
	}

	readFrom := os.Stdin
	if fileArg != "" {
		file, errF := os.Open(filepath.Clean(fileArg))
		if errF != nil {
			return nil, errF
		}

		readFrom = file
	}

	if readFrom == os.Stdin {
		pipeMode, pipeErr := getPipeCharDeviceMode()
		if pipeErr != nil {
			return nil, pipeErr
		}

		if pipeMode != 0 {
			if runtime.GOOS == "windows" {
				fmt.Fprintln(t.Output, breakTextForWindows)
			} else {
				fmt.Fprintln(t.Output, breakTextForUnix)
			}
		}
	}

	return readFrom, nil
}

func (t *Tablo) processHeaders(tw table.Writer, lines []string) []int {
	if len(lines) == 0 || len(t.Args) == 0 {
		return nil
	}

	var headers []string
	var columnIndices []int

	if t.FieldDelimiter == ' ' {
		headers = spaceSplitter(defaultSpaceAmount).Split(lines[0], -1)
	} else {
		repeatAmount := maxConsecutiveRepeats(lines[0], t.FieldDelimiter)
		headers = charSplitter(t.FieldDelimiter, repeatAmount).Split(lines[0], -1)
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
		if len(lines) > 1 {
			tw.AppendHeader(stringSliceToRow(headers))
		} else {
			tw.AppendRow(stringSliceToRow(headers))
		}
	}

	return columnIndices
}

func (t *Tablo) processRows(tw table.Writer, lines []string, columnIndices []int) {
	columnIndicesLen := len(columnIndices)
	filterIndexesLen := len(t.FilterIndexes)

	for i, line := range lines {
		if len(t.Args) > 0 && i == 0 {
			continue
		}
		var fields []string

		if t.FieldDelimiter == ' ' {
			fields = spaceSplitter(defaultSpaceAmount).Split(line, -1)
		} else {
			repeatAmount := maxConsecutiveRepeats(line, t.FieldDelimiter)
			fields = charSplitter(t.FieldDelimiter, repeatAmount).Split(line, -1)
		}

		var selectedFields []string

		switch {
		case filterIndexesLen > 0:
			for _, idx := range t.FilterIndexes {
				if idx >= 0 && idx < len(fields) {
					selectedFields = append(selectedFields, fields[idx])
				} else {
					selectedFields = append(selectedFields, "")
				}
			}

		case columnIndicesLen > 0:
			for _, idx := range columnIndices {
				if idx < len(fields) {
					selectedFields = append(selectedFields, fields[idx])
				} else {
					selectedFields = append(selectedFields, "")
				}
			}

		default:
			selectedFields = fields
		}

		tw.AppendRow(stringSliceToRow(selectedFields))
	}
}

// Tabelize generates tablized output.
func (t *Tablo) Tabelize() error {
	if t.DisplayVersion {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n", t.Version)
		return nil
	}
	readFrom, err := t.getReadFrom()
	if err != nil {
		return err
	}

	defer func() {
		if readFrom != os.Stdin {
			_ = readFrom.Close()
		}
	}()

	input, err := t.ReadInputFunc(readFrom)
	if err != nil {
		return err
	}

	rawLines := strings.FieldsFunc(input, func(r rune) bool {
		return r == t.LineDelimiter
	})

	var lines []string
	for _, line := range rawLines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}

	tw := table.NewWriter()
	tw.SetOutputMirror(t.Output)
	tw.SetStyle(table.StyleLight)
	tw.Style().Format.Header = text.FormatDefault
	tw.Style().Options.SeparateRows = !t.SeparateRows

	headerColumnIndices := t.processHeaders(tw, lines)
	t.processRows(tw, lines, headerColumnIndices)
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
		if output == "" {
			return fmt.Errorf("%w, output can not be an empty string", ErrValueRequired)
		}
		t.Output = os.Stdout
		if output != defaultOutput {
			f, err := os.Create(filepath.Clean(output))
			if err != nil {
				return fmt.Errorf("%w, %w", ErrInvalidValue, err)
			}
			t.Output = f
		}

		return nil
	}
}

// WithOutputWriter sets the output write for test usage.
func WithOutputWriter(wr io.WriteCloser) Option {
	return func(t *Tablo) error {
		if wr == nil {
			return fmt.Errorf("%w, output writer is nil", ErrValueRequired)
		}
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

// WithLineDelimiter sets the line delimiter.
func WithLineDelimiter(s string) Option {
	return func(t *Tablo) error {
		if s == "" {
			return fmt.Errorf("%w, line delimiter is empty string", ErrValueRequired)
		}

		var delimiter rune

		switch s {
		case `\n`:
			delimiter = defaultLineDelimiter
		case `\r`:
			delimiter = '\r'
		case `\t`:
			delimiter = '\t'
		default:
			delimiter = []rune(s)[0]
		}

		t.LineDelimiter = delimiter

		return nil
	}
}

// WithFieldDelimiter sets the field delimiter.
func WithFieldDelimiter(s string) Option {
	return func(t *Tablo) error {
		if s == "" {
			t.FieldDelimiter = 0

			return nil
		}

		var delimiter rune

		switch s {
		case `\f`:
			delimiter = '\f'
		case `\v`:
			delimiter = '\v'
		case `\t`:
			delimiter = '\t'
		default:
			delimiter = []rune(s)[0]
		}

		t.FieldDelimiter = delimiter

		return nil
	}
}

// WithNoSeparateRows disables the separation row line.
func WithNoSeparateRows(sep bool) Option {
	return func(t *Tablo) error {
		t.SeparateRows = sep

		return nil
	}
}

// WithFilterIndexes sets the filter index columns.
func WithFilterIndexes(indexes string) Option {
	return func(t *Tablo) error {
		if indexes == "" {
			return nil
		}

		ss := strings.Split(indexes, ",")
		if len(ss) > 0 {
			idxes := make([]int, len(ss))
			for i, v := range ss {
				n, err := strconv.Atoi(v)
				if err != nil {
					return fmt.Errorf("%w, %s is not a number", ErrInvalidValue, v)
				}

				if n > 0 {
					idxes[i] = n - 1
				}

			}

			t.FilterIndexes = idxes
		}

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

// Run runs the command.
func Run() error {
	flag.Usage = getUsage
	output := flag.String("output", defaultOutput, helpOutput)
	flag.StringVar(output, "o", defaultOutput, helpOutput+" (short)")

	version := flag.Bool("version", false, "display version information")

	lineDelimiterChar := flag.String("line-delimiter-char", string(defaultLineDelimiter), helpLineDelimiterChar)
	flag.StringVar(lineDelimiterChar, "l", string(defaultLineDelimiter), helpLineDelimiterChar+" (short)")

	fieldDelimiterChar := flag.String("field-delimiter-char", string(defaultFieldDelimiter), helpFieldDelimiterChar)
	flag.StringVar(fieldDelimiterChar, "f", string(defaultFieldDelimiter), helpFieldDelimiterChar+" (short)")

	noSeparateRows := flag.Bool("no-separate-rows", false, helpNoSeparateRows)
	flag.BoolVar(noSeparateRows, "n", false, helpNoSeparateRows+" (short)")

	filterIndexes := flag.String("filter-indexes", "", helpFilterIndexes)
	flag.StringVar(filterIndexes, "fi", "", helpFilterIndexes+" (short)")

	flag.Parse()

	tbl, err := New(
		WithArgs(flag.Args()),
		WithOutput(*output),
		WithDisplayVersion(*version),
		WithReadInputFunc(readInput),
		WithLineDelimiter(*lineDelimiterChar),
		WithFieldDelimiter(*fieldDelimiterChar),
		WithNoSeparateRows(*noSeparateRows),
		WithFilterIndexes(*filterIndexes),
	)
	if err != nil {
		return err
	}

	if err = tbl.Tabelize(); err != nil {
		return err
	}

	if *output != defaultOutput {
		fmt.Fprintf(flag.CommandLine.Output(), "result saved to: %s\n", *output)
		defer func() { _ = tbl.Output.Close() }()
	}

	return nil
}
