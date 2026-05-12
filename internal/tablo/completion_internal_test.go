package tablo

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingWriter struct {
	err error
}

func (f failingWriter) Write([]byte) (int, error) {
	return 0, f.err
}

func TestBashCompletionScript(t *testing.T) {
	script := bashCompletionScript("tablo")

	assert.Contains(t, script, "_tablo_completion() {")
	assert.Contains(t, script, "complete -F _tablo_completion -- 'tablo'")
	assert.Contains(t, script, completeFlag)
	assert.Contains(t, script, "positional_count=0")
	assert.Contains(t, script, "saw_double_dash=0")
	assert.Contains(t, script, "if (( saw_double_dash == 0 )); then")
	assert.Contains(t, script, "if (( saw_double_dash == 1 )); then")
	assert.Contains(t, script, "-fi|-filter-indexes|--filter-indexes")
	assert.Contains(t, script, "-field-delimiter-char")
	assert.Contains(t, script, "-o=*|-output=*|--output=*")
	assert.NotContains(t, script, "mapfile")
	assert.NotContains(t, script, "compopt")
}

func TestSanitizeCompletionFunctionName(t *testing.T) {
	assert.Equal(t, "tablo_dev", sanitizeCompletionFunctionName("tablo-dev"))
	assert.Equal(t, "_1tablo", sanitizeCompletionFunctionName("1tablo"))
	assert.Equal(t, "tablo", sanitizeCompletionFunctionName(""))
}

func TestBashCompletionScript_UsesSanitizedFunctionName(t *testing.T) {
	script := bashCompletionScript("tablo-dev")

	assert.Contains(t, script, "_tablo_dev_completion() {")
	assert.Contains(t, script, "complete -F _tablo_dev_completion -- 'tablo-dev'")
	assert.NotContains(t, script, "_tablo-dev_completion() {")
}

func TestShellQuote(t *testing.T) {
	assert.Equal(t, "'tablo dev'", shellQuote("tablo dev"))
	assert.Equal(t, `'tablo'\''dev'`, shellQuote("tablo'dev"))
}

func TestResolveCompletionPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	t.Setenv("TABLO_TEST_DIR", "tablo")

	path, err := resolveCompletionPath("~/data.csv")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(homeDir, "data.csv"), path)

	path, err = resolveCompletionPath("$TABLO_TEST_DIR/data.csv")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("tablo", "data.csv"), path)
}

func TestCompletionSuggestions_Flags(t *testing.T) {
	suggestions, err := completionSuggestions([]string{"tablo", "--j"}, 1)

	require.NoError(t, err)
	assert.Equal(t, []string{"--json"}, suggestions)
}

func TestCompletionSuggestions_AllFlagsForFirstArgument(t *testing.T) {
	suggestions, err := completionSuggestions([]string{"tablo", ""}, 1)

	require.NoError(t, err)
	assert.Contains(t, suggestions, shortBashCompletionFlag)
	assert.Contains(t, suggestions, bashCompletionFlag)
	assert.Contains(t, suggestions, "-h")
	assert.Contains(t, suggestions, "--help")
	assert.Contains(t, suggestions, "--output")
	assert.Contains(t, suggestions, "-f")
}

func TestCompletionSuggestions_FieldDelimiterValues(t *testing.T) {
	suggestions, err := completionSuggestions([]string{"tablo", "-f", ":"}, 2)

	require.NoError(t, err)
	assert.Contains(t, suggestions, ":")
	assert.NotContains(t, suggestions, " ")
}

func TestCompletionSuggestions_LineDelimiterValues(t *testing.T) {
	suggestions, err := completionSuggestions([]string{"tablo", "-l", "\\r"}, 2)

	require.NoError(t, err)
	assert.Equal(t, []string{"\\r"}, suggestions)
}

func TestCompletionSuggestions_LongFieldDelimiterValues(t *testing.T) {
	suggestions, err := completionSuggestions([]string{"tablo", "-field-delimiter-char", ":"}, 2)

	require.NoError(t, err)
	assert.Equal(t, []string{":"}, suggestions)
}

func TestCompletionSuggestions_InlineFieldDelimiterValues(t *testing.T) {
	suggestions, err := completionSuggestions([]string{"tablo", "--field-delimiter-char=:"}, 1)

	require.NoError(t, err)
	assert.Equal(t, []string{"--field-delimiter-char=:"}, suggestions)
}

func TestCompletionSuggestions_InlineLineDelimiterValues(t *testing.T) {
	suggestions, err := completionSuggestions([]string{"tablo", "--line-delimiter-char=\\r"}, 1)

	require.NoError(t, err)
	assert.Equal(t, []string{"--line-delimiter-char=\\r"}, suggestions)
}

func TestCompletionSuggestions_NoColumnCompletionWhenFilterIndexesSet(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "users.csv")
	err := os.WriteFile(inputFile, []byte("Username;Identifier\nbooker12;9012\n"), 0o600)
	require.NoError(t, err)

	suggestions, err := completionSuggestions([]string{"tablo", "-fi", "1,2", inputFile, ""}, 4)

	require.NoError(t, err)
	assert.Nil(t, suggestions)
}

func TestCompletionSuggestions_NoColumnCompletionWhenFirstPositionalIsNotFile(t *testing.T) {
	suggestions, err := completionSuggestions([]string{"tablo", "Username"}, 1)

	require.NoError(t, err)
	assert.Nil(t, suggestions)
}

func TestCompletionSuggestions_AfterDoubleDashDoesNotSuggestFlags(t *testing.T) {
	suggestions, err := completionSuggestions([]string{"tablo", "--", "-weird.csv"}, 2)

	require.NoError(t, err)
	assert.Nil(t, suggestions)
}

func TestCompletionSuggestions_AfterDoubleDashDoesNotSuggestFlagValues(t *testing.T) {
	suggestions, err := completionSuggestions([]string{"tablo", "--", "-f"}, 2)

	require.NoError(t, err)
	assert.Nil(t, suggestions)
}

func TestCompletionSuggestions_AfterPositionalDoesNotSuggestFlags(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "users.csv")
	err := os.WriteFile(inputFile, []byte("Username;Identifier\nbooker12;9012\n"), 0o600)
	require.NoError(t, err)

	suggestions, err := completionSuggestions([]string{"tablo", inputFile, "-n"}, 2)

	require.NoError(t, err)
	assert.Nil(t, suggestions)
}

func TestCompletionSuggestions_AfterDoubleDashDoesNotSuggestInlineFlagValues(t *testing.T) {
	suggestions, err := completionSuggestions([]string{"tablo", "--", "--field-delimiter-char=:"}, 2)

	require.NoError(t, err)
	assert.Nil(t, suggestions)
}

func TestCompletionSuggestions_ColumnsFromInputFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "users.csv")
	err := os.WriteFile(inputFile, []byte("Username;Identifier;First name\nbooker12;9012;Rachel\n"), 0o600)
	require.NoError(t, err)

	suggestions, err := completionSuggestions([]string{"tablo", "-f", ";", inputFile, "Fi"}, 4)

	require.NoError(t, err)
	assert.Equal(t, []string{"First name"}, suggestions)
}

func TestCompletionSuggestions_ColumnsFromInputFile_AutoDetectDelimiter(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "users.csv")
	err := os.WriteFile(inputFile, []byte("Username,Identifier,First name\nbooker12,9012,Rachel\n"), 0o600)
	require.NoError(t, err)

	suggestions, err := completionSuggestions([]string{"tablo", inputFile, "Id"}, 2)

	require.NoError(t, err)
	assert.Equal(t, []string{"Identifier"}, suggestions)
}

func TestCompletionSuggestions_ColumnsExcludeAlreadySelected(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "users.csv")
	err := os.WriteFile(inputFile, []byte("Username;Identifier;First name\nbooker12;9012;Rachel\n"), 0o600)
	require.NoError(t, err)

	suggestions, err := completionSuggestions([]string{"tablo", "-f", ";", inputFile, "Username", ""}, 5)

	require.NoError(t, err)
	assert.Equal(t, []string{"Identifier", "First name"}, suggestions)
}

func TestRunCompletion_NoWords(t *testing.T) {
	var output bytes.Buffer

	err := runCompletion(nil, &output)

	require.NoError(t, err)
	assert.Empty(t, output.String())
}

func TestRunCompletion_WritesSuggestions(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "users.csv")
	err := os.WriteFile(inputFile, []byte("Username;Identifier\nbooker12;9012\n"), 0o600)
	require.NoError(t, err)

	t.Setenv("COMP_CWORD", "4")

	var output bytes.Buffer
	err = runCompletion([]string{"--", "tablo", "-f", ";", inputFile, "Us"}, &output)

	require.NoError(t, err)
	assert.Equal(t, "Username", strings.TrimSpace(output.String()))
}

func TestRunCompletion_InvalidCompCwordFallsBack(t *testing.T) {
	t.Setenv("COMP_CWORD", "invalid")

	var output bytes.Buffer
	err := runCompletion([]string{"--", "tablo", "--j"}, &output)

	require.NoError(t, err)
	assert.Equal(t, "--json", strings.TrimSpace(output.String()))
}

func TestRunCompletion_ReturnsWriteError(t *testing.T) {
	writeErr := errors.New("write failed")
	t.Setenv("COMP_CWORD", "2")

	err := runCompletion([]string{"--", "tablo", "--j"}, failingWriter{err: writeErr})

	assert.ErrorIs(t, err, writeErr)
}

func TestParseCompletionState(t *testing.T) {
	state := completionState{
		lineDelimiter: defaultLineDelimiter,
	}

	err := parseCompletionState([]string{"tablo", "--field-delimiter-char=;", "-l", "\\t", "-fi", "1,2", "--json", "users.csv", "Username"}, 8, &state)

	require.NoError(t, err)
	assert.Equal(t, ';', state.fieldDelimiter)
	assert.Equal(t, '\t', state.lineDelimiter)
	assert.True(t, state.filterIndexes)
	assert.Equal(t, []string{"users.csv"}, state.positionals)
}

func TestParseCompletionState_SingleDashLongFlags(t *testing.T) {
	state := completionState{
		lineDelimiter: defaultLineDelimiter,
	}

	err := parseCompletionState([]string{"tablo", "-field-delimiter-char=;", "-line-delimiter-char", "\\t", "-filter-indexes", "1,2", "-json", "users.csv"}, 7, &state)

	require.NoError(t, err)
	assert.Equal(t, ';', state.fieldDelimiter)
	assert.Equal(t, '\t', state.lineDelimiter)
	assert.True(t, state.filterIndexes)
	assert.Empty(t, state.positionals)
}

func TestParseCompletionState_EndOfFlags(t *testing.T) {
	state := completionState{
		lineDelimiter: defaultLineDelimiter,
	}

	err := parseCompletionState([]string{"tablo", "--", "users.csv", "Username"}, 4, &state)

	require.NoError(t, err)
	assert.Equal(t, []string{"users.csv", "Username"}, state.positionals)
}

func TestCompletionFlagToken(t *testing.T) {
	flagName, flagValue, hasInlineValue := completionFlagToken("--output=file.txt")
	assert.Equal(t, "--output", flagName)
	assert.Equal(t, "file.txt", flagValue)
	assert.True(t, hasInlineValue)

	flagName, flagValue, hasInlineValue = completionFlagToken("--json")
	assert.Equal(t, "--json", flagName)
	assert.Empty(t, flagValue)
	assert.False(t, hasInlineValue)
}

func TestApplyCompletionFlagValue(t *testing.T) {
	state := completionState{
		lineDelimiter: defaultLineDelimiter,
	}

	applyCompletionFlagValue(&state, "--line-delimiter-char", "\\r")
	applyCompletionFlagValue(&state, "--field-delimiter-char", "|")
	applyCompletionFlagValue(&state, "--filter-indexes", "1")
	applyCompletionFlagValue(&state, "--filter-indexes", "")

	assert.Equal(t, '\r', state.lineDelimiter)
	assert.Equal(t, '|', state.fieldDelimiter)
	assert.False(t, state.filterIndexes)
}

func TestApplyCompletionFlagValue_SingleDashLongFlags(t *testing.T) {
	state := completionState{
		lineDelimiter: defaultLineDelimiter,
	}

	applyCompletionFlagValue(&state, "-line-delimiter-char", "\\r")
	applyCompletionFlagValue(&state, "-field-delimiter-char", ";")
	applyCompletionFlagValue(&state, "-filter-indexes", "2")

	assert.Equal(t, '\r', state.lineDelimiter)
	assert.Equal(t, ';', state.fieldDelimiter)
	assert.True(t, state.filterIndexes)
}

func TestCompletionValueSuggestions_UnknownFlag(t *testing.T) {
	assert.Nil(t, completionValueSuggestions("--unknown", ""))
}

func TestCompletionInlineValueSuggestions(t *testing.T) {
	assert.Equal(
		t,
		[]string{"--field-delimiter-char=:"},
		completionInlineValueSuggestions("--field-delimiter-char=:"),
	)
	assert.Nil(t, completionInlineValueSuggestions("--output=foo"))
	assert.Nil(t, completionInlineValueSuggestions("--json"))
}

func TestCompletionPrefixMatches(t *testing.T) {
	assert.Equal(t, []string{"--json"}, completionPrefixMatches([]string{"--json", "--output"}, "--J"))
	assert.Equal(t, []string{"a", "b"}, completionPrefixMatches([]string{"a", "b"}, ""))
}

func TestCurrentCompletionWord_OutOfRange(t *testing.T) {
	assert.Empty(t, currentCompletionWord([]string{"tablo"}, 4))
}

func TestIsRegularFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "users.csv")
	err := os.WriteFile(inputFile, []byte("Username\n"), 0o600)
	require.NoError(t, err)

	assert.True(t, isRegularFile(inputFile))
	assert.False(t, isRegularFile(tmpDir))
	assert.False(t, isRegularFile(filepath.Join(tmpDir, "missing.csv")))
}

func TestCompleteColumnsFromFile_CommentOnlyInput(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "users.csv")
	err := os.WriteFile(inputFile, []byte("# comment\n"), 0o600)
	require.NoError(t, err)

	suggestions, err := completeColumnsFromFile(completionState{lineDelimiter: '\n'}, inputFile, nil, "")

	require.NoError(t, err)
	assert.Nil(t, suggestions)
}

func TestCompleteColumnsFromFile_ReadsOnlyHeaderProbeLines(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "users.csv")
	content := strings.Join([]string{
		"# comment",
		"",
		"Username,Identifier,First name",
		"booker12,9012,Rachel",
		"grey07,2070,Laura",
		"johnson81,4081,Craig",
		"jenkins46,9346,Mary",
		"smith79,5079,Jamie",
	}, "\n")
	err := os.WriteFile(inputFile, []byte(content), 0o600)
	require.NoError(t, err)

	suggestions, err := completeColumnsFromFile(completionState{lineDelimiter: '\n'}, inputFile, nil, "Fi")

	require.NoError(t, err)
	assert.Equal(t, []string{"First name"}, suggestions)
}

func TestCompleteColumnsFromFile_WithoutHeaderReturnsNoSuggestions(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "users.csv")
	content := strings.Join([]string{
		"booker12;9012;Rachel",
		"grey07;2070;Laura",
	}, "\n")
	err := os.WriteFile(inputFile, []byte(content), 0o600)
	require.NoError(t, err)

	suggestions, err := completeColumnsFromFile(completionState{
		lineDelimiter:  '\n',
		fieldDelimiter: ';',
	}, inputFile, nil, "Ra")

	require.NoError(t, err)
	assert.Nil(t, suggestions)
}

func TestCompleteColumnsFromFile_ResolvesHomePath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	inputFile := filepath.Join(homeDir, "tablo-completion-home.csv")
	err = os.WriteFile(inputFile, []byte("Username;Identifier\nbooker12;9012\n"), 0o600)
	require.NoError(t, err)
	defer func() { _ = os.Remove(inputFile) }()

	suggestions, err := completeColumnsFromFile(completionState{
		lineDelimiter:  '\n',
		fieldDelimiter: ';',
	}, "~/tablo-completion-home.csv", nil, "Us")

	require.NoError(t, err)
	assert.Equal(t, []string{"Username"}, suggestions)
}

func TestReadCompletionLines(t *testing.T) {
	lines, err := readCompletionLines(strings.NewReader("# comment\n\nname,age\nvigo,42\nignored,after\n"), '\n', 2)

	require.NoError(t, err)
	assert.Equal(t, []string{"name,age", "vigo,42"}, lines)
}

func TestReadCompletionLines_StopsAtByteCap(t *testing.T) {
	veryLong := strings.Repeat("a", completionByteCap+10)

	lines, err := readCompletionLines(strings.NewReader(veryLong), '\n', 2)

	require.NoError(t, err)
	require.Len(t, lines, 1)
	assert.Len(t, lines[0], completionByteCap)
}

func TestCompleteColumnsFromFile_OpenError(t *testing.T) {
	_, err := completeColumnsFromFile(completionState{lineDelimiter: '\n'}, "/does/not/exist.csv", nil, "")

	assert.Error(t, err)
}

func TestRun_CompleteFlag(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "users.csv")
	err := os.WriteFile(inputFile, []byte("Username;Identifier\nbooker12;9012\n"), 0o600)
	require.NoError(t, err)

	t.Setenv("COMP_CWORD", "4")
	oldArgs := os.Args
	oldFlagSet := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlagSet
	}()
	os.Args = []string{"tablo", "--complete", "--", "tablo", "-f", ";", inputFile, "Us"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	err = Run()
	require.NoError(t, err)
	_ = w.Close()

	output, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, "Username\n", string(output))
}
