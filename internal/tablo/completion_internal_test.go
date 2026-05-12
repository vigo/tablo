package tablo

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBashCompletionScript(t *testing.T) {
	script := bashCompletionScript("tablo")

	assert.Contains(t, script, "complete -F _tablo_completion tablo")
	assert.Contains(t, script, completeFlag)
}

func TestSanitizeCompletionFunctionName(t *testing.T) {
	assert.Equal(t, "tablo_dev", sanitizeCompletionFunctionName("tablo-dev"))
}

func TestCompletionSuggestions_Flags(t *testing.T) {
	suggestions, err := completionSuggestions([]string{"tablo", "--j"}, 1)

	require.NoError(t, err)
	assert.Equal(t, []string{"--json"}, suggestions)
}

func TestCompletionSuggestions_FieldDelimiterValues(t *testing.T) {
	suggestions, err := completionSuggestions([]string{"tablo", "-f", ":"}, 2)

	require.NoError(t, err)
	assert.Contains(t, suggestions, ":")
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

func TestCompletionSuggestions_ColumnsExcludeAlreadySelected(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "users.csv")
	err := os.WriteFile(inputFile, []byte("Username;Identifier;First name\nbooker12;9012;Rachel\n"), 0o600)
	require.NoError(t, err)

	suggestions, err := completionSuggestions([]string{"tablo", "-f", ";", inputFile, "Username", ""}, 5)

	require.NoError(t, err)
	assert.Equal(t, []string{"Identifier", "First name"}, suggestions)
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
