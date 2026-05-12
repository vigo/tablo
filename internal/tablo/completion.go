package tablo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	shortBashCompletionFlag = "-bash-completion"
	bashCompletionFlag      = "--bash-completion"
	completeFlag            = "--complete"
	completionByteCap       = 64 * 1024
)

var (
	completionBooleanFlags = map[string]struct{}{
		shortBashCompletionFlag: {},
		bashCompletionFlag:      {},
		"-h":                    {},
		"-help":                 {},
		"--help":                {},
		"-version":              {},
		"--version":             {},
		"-n":                    {},
		"-no-separate-rows":     {},
		"--no-separate-rows":    {},
		"-nb":                   {},
		"-no-borders":           {},
		"--no-borders":          {},
		"-nh":                   {},
		"-no-headers":           {},
		"--no-headers":          {},
		"-j":                    {},
		"-json":                 {},
		"--json":                {},
	}
	completionValueFlags = map[string]struct{}{
		"-f":                     {},
		"-field-delimiter-char":  {},
		"--field-delimiter-char": {},
		"-l":                     {},
		"-line-delimiter-char":   {},
		"--line-delimiter-char":  {},
		"-fi":                    {},
		"-filter-indexes":        {},
		"--filter-indexes":       {},
		"-o":                     {},
		"-output":                {},
		"--output":               {},
	}
	completionAllFlags = []string{
		shortBashCompletionFlag,
		bashCompletionFlag,
		"-h",
		"-help",
		"--help",
		"-version",
		"--version",
		"-f",
		"-field-delimiter-char",
		"--field-delimiter-char",
		"-l",
		"-line-delimiter-char",
		"--line-delimiter-char",
		"-n",
		"-no-separate-rows",
		"--no-separate-rows",
		"-nb",
		"-no-borders",
		"--no-borders",
		"-nh",
		"-no-headers",
		"--no-headers",
		"-fi",
		"-filter-indexes",
		"--filter-indexes",
		"-j",
		"-json",
		"--json",
		"-o",
		"-output",
		"--output",
	}
)

type completionState struct {
	fieldDelimiter rune
	lineDelimiter  rune
	filterIndexes  bool
	positionals    []string
}

func bashCompletionScript(binaryName string) string {
	functionName := sanitizeCompletionFunctionName(binaryName)
	quotedBinaryName := shellQuote(binaryName)

	return fmt.Sprintf(`_%[1]s_completion() {
    local cur prev word expect_value positional_count reply prefix value saw_double_dash i
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev=""
    expect_value=0
    positional_count=0
    saw_double_dash=0
    if (( COMP_CWORD > 0 )); then
        prev="${COMP_WORDS[COMP_CWORD-1]}"
    fi

    for (( i=1; i<COMP_CWORD; i++ )); do
        word="${COMP_WORDS[i]}"

        if (( expect_value == 1 )); then
            expect_value=0
            continue
        fi
        if (( saw_double_dash == 1 )); then
            positional_count=$(( positional_count + 1 ))
            continue
        fi

        case "${word}" in
            --)
                saw_double_dash=1
                continue
                ;;
            -f|-field-delimiter-char|--field-delimiter-char|\
            -l|-line-delimiter-char|--line-delimiter-char|\
            -fi|-filter-indexes|--filter-indexes|\
            -o|-output|--output)
                expect_value=1
                continue
                ;;
            -field-delimiter-char=*|--field-delimiter-char=*|\
            -line-delimiter-char=*|--line-delimiter-char=*|\
            -filter-indexes=*|--filter-indexes=*|\
            -output=*|--output=*)
                continue
                ;;
            -h|-help|--help|\
            -version|--version|\
            -n|-no-separate-rows|--no-separate-rows|\
            -nb|-no-borders|--no-borders|\
            -nh|-no-headers|--no-headers|\
            -j|-json|--json)
                continue
                ;;
            -*)
                continue
                ;;
        esac

        positional_count=$(( positional_count + 1 ))
    done

    if (( saw_double_dash == 0 )); then
        case "${prev}" in
            -o|-output|--output)
                while IFS= read -r reply; do
                    COMPREPLY+=("${reply}")
                done < <(compgen -f -- "${cur}")
                return 0
                ;;
        esac

        case "${cur}" in
            -o=*|-output=*|--output=*)
                prefix="${cur%%=*}="
                value="${cur#*=}"
                while IFS= read -r reply; do
                    COMPREPLY+=("${prefix}${reply}")
                done < <(compgen -f -- "${value}")
                return 0
                ;;
        esac
    fi

    while IFS= read -r reply; do
        COMPREPLY+=("${reply}")
    done < <(COMP_CWORD="${COMP_CWORD}" "${COMP_WORDS[0]}" %[2]s -- "${COMP_WORDS[@]}")

    if (( ${#COMPREPLY[@]} > 0 )); then
        return 0
    fi

    case "${prev}" in
        -f|-field-delimiter-char|--field-delimiter-char|\
        -l|-line-delimiter-char|--line-delimiter-char|\
        -fi|-filter-indexes|--filter-indexes)
            return 0
            ;;
    esac

    if (( positional_count == 0 )) && { [[ "${cur}" != -* ]] || (( saw_double_dash == 1 )); }; then
        while IFS= read -r reply; do
            COMPREPLY+=("${reply}")
        done < <(compgen -f -- "${cur}")
    fi
}

complete -F _%[1]s_completion -- %[3]s
`, functionName, completeFlag, quotedBinaryName)
}

func runCompletion(words []string, output io.Writer) error {
	if len(words) > 0 && words[0] == "--" {
		words = words[1:]
	}
	if len(words) == 0 {
		return nil
	}

	cword, err := strconv.Atoi(os.Getenv("COMP_CWORD"))
	if err != nil || cword < 0 {
		cword = len(words) - 1
	}
	if cword >= len(words) {
		cword = len(words) - 1
	}

	suggestions, err := completionSuggestions(words, cword)
	if err != nil {
		return err
	}

	for _, suggestion := range suggestions {
		if _, err := fmt.Fprintln(output, suggestion); err != nil {
			return fmt.Errorf(errorWrapFormat, err)
		}
	}

	return nil
}

func completionSuggestions(words []string, cword int) ([]string, error) {
	if len(words) == 0 || cword <= 0 {
		return completionFlagMatches(currentCompletionWord(words, cword)), nil
	}

	current := currentCompletionWord(words, cword)
	afterDoubleDash := completionAfterDoubleDash(words, cword)
	if !afterDoubleDash {
		if suggestions := completionInlineValueSuggestions(current); suggestions != nil {
			return suggestions, nil
		}

		previous := words[cword-1]
		if suggestions := completionValueSuggestions(previous, current); suggestions != nil {
			return suggestions, nil
		}
	}

	state := completionState{
		lineDelimiter: defaultLineDelimiter,
	}
	if err := parseCompletionState(words, cword, &state); err != nil {
		return nil, err
	}

	if !afterDoubleDash && (strings.HasPrefix(current, "-") || current == "" && cword == 1) {
		return completionFlagMatches(current), nil
	}
	if state.filterIndexes || len(state.positionals) == 0 {
		return nil, nil
	}
	if !isRegularFile(state.positionals[0]) {
		return nil, nil
	}

	return completeColumnsFromFile(state, state.positionals[0], state.positionals[1:], current)
}

func completionAfterDoubleDash(words []string, cword int) bool {
	for i := 1; i < cword && i < len(words); i++ {
		if words[i] == "--" {
			return true
		}
	}

	return false
}

func parseCompletionState(words []string, cword int, state *completionState) error {
	expectingValue := ""

	for i := 1; i < cword; i++ {
		token := words[i]
		if expectingValue != "" {
			applyCompletionFlagValue(state, expectingValue, token)
			expectingValue = ""
			continue
		}
		if token == "--" {
			state.positionals = append(state.positionals, words[i+1:cword]...)
			break
		}

		flagName, flagValue, hasInlineValue := completionFlagToken(token)
		switch {
		case completionHasValueFlag(flagName):
			if hasInlineValue {
				applyCompletionFlagValue(state, flagName, flagValue)
			} else {
				expectingValue = flagName
			}
		case completionHasBooleanFlag(flagName):
			continue
		default:
			state.positionals = append(state.positionals, token)
		}
	}

	return nil
}

func completionFlagToken(token string) (flagName, flagValue string, hasInlineValue bool) {
	if head, tail, ok := strings.Cut(token, "="); ok {
		return head, tail, true
	}

	return token, "", false
}

func completionHasBooleanFlag(flagName string) bool {
	_, ok := completionBooleanFlags[flagName]
	return ok
}

func completionHasValueFlag(flagName string) bool {
	_, ok := completionValueFlags[flagName]
	return ok
}

func applyCompletionFlagValue(state *completionState, flagName, value string) {
	switch flagName {
	case "-f", "-field-delimiter-char", "--field-delimiter-char":
		if value != "" {
			state.fieldDelimiter = parseSpecialChars(value)
		}
	case "-l", "-line-delimiter-char", "--line-delimiter-char":
		if value != "" {
			state.lineDelimiter = parseSpecialChars(value)
		}
	case "-fi", "-filter-indexes", "--filter-indexes":
		state.filterIndexes = value != ""
	}
}

func completionValueSuggestions(flagName, current string) []string {
	switch flagName {
	case "-f", "-field-delimiter-char", "--field-delimiter-char":
		return completionPrefixMatches([]string{",", ";", "|", ":", "\\t"}, current)
	case "-l", "-line-delimiter-char", "--line-delimiter-char":
		return completionPrefixMatches([]string{"\\n", "\\t", "\\r", ":", ";", "|"}, current)
	default:
		return nil
	}
}

func completionFlagMatches(current string) []string {
	return completionPrefixMatches(completionAllFlags, current)
}

func completionInlineValueSuggestions(current string) []string {
	flagName, currentValue, hasInlineValue := completionFlagToken(current)
	if !hasInlineValue {
		return nil
	}

	if flagName == "-o" || flagName == "-output" || flagName == "--output" {
		return nil
	}

	suggestions := completionValueSuggestions(flagName, currentValue)
	if suggestions == nil {
		return nil
	}

	prefixed := make([]string, 0, len(suggestions))
	for _, suggestion := range suggestions {
		prefixed = append(prefixed, flagName+"="+suggestion)
	}

	return prefixed
}

func completionPrefixMatches(candidates []string, current string) []string {
	if current == "" {
		return append([]string(nil), candidates...)
	}

	var suggestions []string
	for _, candidate := range candidates {
		if strings.HasPrefix(strings.ToLower(candidate), strings.ToLower(current)) {
			suggestions = append(suggestions, candidate)
		}
	}

	return suggestions
}

func currentCompletionWord(words []string, cword int) string {
	if cword >= 0 && cword < len(words) {
		return words[cword]
	}

	return ""
}

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.Mode().IsRegular()
}

func sanitizeCompletionFunctionName(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}

	if b.Len() == 0 {
		return "tablo"
	}

	sanitized := b.String()
	first, _ := utf8.DecodeRuneInString(sanitized)
	if unicode.IsDigit(first) {
		return "_" + sanitized
	}

	return sanitized
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func completeColumnsFromFile(state completionState, path string, selected []string, current string) ([]string, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf(errorWrapFormat, err)
	}
	defer func() { _ = file.Close() }()

	lines, err := readCompletionLines(file, state.lineDelimiter, delimiterProbeLines)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, nil
	}

	tbl := &Tablo{
		FieldDelimiter: state.fieldDelimiter,
	}
	tbl.ensureDetectedFieldDelimiter(lines)

	headers := tbl.splitFields(lines[0])
	seen := make(map[string]struct{}, len(selected))
	for _, column := range selected {
		seen[strings.ToLower(column)] = struct{}{}
	}

	var suggestions []string
	for _, header := range headers {
		if _, ok := seen[strings.ToLower(header)]; ok {
			continue
		}
		if current == "" || strings.HasPrefix(strings.ToLower(header), strings.ToLower(current)) {
			suggestions = append(suggestions, header)
		}
	}

	return suggestions, nil
}

func readCompletionLines(reader io.Reader, delimiter rune, limit int) ([]string, error) {
	if limit <= 0 {
		return nil, nil
	}

	buffered := bufio.NewReader(reader)
	lines := make([]string, 0, limit)
	var currentLine strings.Builder
	bytesRead := 0

	flushLine := func() {
		line := currentLine.String()
		currentLine.Reset()
		if delimiter == '\n' {
			line = strings.TrimSuffix(line, "\r")
		}

		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			return
		}

		lines = append(lines, line)
	}

	for len(lines) < limit {
		r, size, err := buffered.ReadRune()
		switch err {
		case nil:
			bytesRead += size
			if bytesRead > completionByteCap {
				if currentLine.Len() > 0 {
					flushLine()
				}
				return lines, nil
			}
			if r == delimiter {
				flushLine()
				continue
			}
			currentLine.WriteRune(r)
		case io.EOF:
			if currentLine.Len() > 0 {
				flushLine()
			}
			return lines, nil
		default:
			return nil, fmt.Errorf(errorWrapFormat, err)
		}
	}

	return lines, nil
}
