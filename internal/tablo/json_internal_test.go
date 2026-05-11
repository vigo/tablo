package tablo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errorWriteCloser struct {
	err error
}

func (e *errorWriteCloser) Write([]byte) (int, error) {
	return 0, e.err
}

func (e *errorWriteCloser) Close() error {
	return nil
}

func TestPickFieldsByIndices_OutOfRange(t *testing.T) {
	fields := pickFieldsByIndices([]string{"foo"}, []int{0, 1})

	assert.Equal(t, []string{"foo", ""}, fields)
}

func TestTablo_SelectFields_ColumnIndices_OutOfRange(t *testing.T) {
	tbl := &Tablo{}

	fields := tbl.selectFields([]string{"foo"}, []int{0, 1})

	assert.Equal(t, []string{"foo", ""}, fields)
}

func TestTablo_SelectFields_FilterIndexes_OutOfRange(t *testing.T) {
	tbl := &Tablo{
		FilterIndexes: []int{1, 2},
	}

	fields := tbl.selectFields([]string{"foo", "bar"}, nil)

	assert.Equal(t, []string{"bar", ""}, fields)
}

func TestTablo_BuildJSONDataset_EmptyLines(t *testing.T) {
	tbl := &Tablo{}

	dataset := tbl.buildJSONDataset(nil)

	assert.Equal(t, jsonDataset{rows: [][]string{}}, dataset)
}

func TestTablo_BuildJSONDataset_WithSelectedHeaders(t *testing.T) {
	tbl := &Tablo{
		Args:           []string{"name", "age"},
		FieldDelimiter: '|',
	}

	dataset := tbl.buildJSONDataset([]string{"name|age|city", "vigo"})

	assert.True(t, dataset.hasHeader)
	assert.Equal(t, []string{"name", "age"}, dataset.headers)
	assert.Equal(t, [][]string{{"vigo", ""}}, dataset.rows)
}

func TestLooksLikeHeader(t *testing.T) {
	assert.True(t, looksLikeHeader([]string{"name", "age"}))
	assert.True(t, looksLikeHeader([]string{"ID", `"T.C. KİMLİK NO"`, "-BÖLÜM", `"_ADI SOYADI"`}))
	assert.False(t, looksLikeHeader([]string{"nobody", "*", "/usr/bin/false"}))
	assert.False(t, looksLikeHeader([]string{"single value"}))
}

func TestTablo_DetectFieldDelimiter(t *testing.T) {
	tbl := &Tablo{}

	delimiter := tbl.detectFieldDelimiter([]string{
		`ID,"name",department`,
		`1,"vigo",engineering`,
	})

	assert.Equal(t, ',', delimiter)
}

func TestTablo_DetectFieldDelimiter_UsesConfiguredDelimiter(t *testing.T) {
	tbl := &Tablo{
		FieldDelimiter: ';',
	}

	delimiter := tbl.detectFieldDelimiter([]string{"name,age", "vigo,42"})

	assert.Equal(t, ';', delimiter)
}

func TestTablo_DetectFieldDelimiter_DoesNotAssumeQuotedCSVParsing(t *testing.T) {
	tbl := &Tablo{}

	delimiter := tbl.detectFieldDelimiter([]string{
		`name,notes`,
		`vigo,"a,b"`,
	})

	assert.Equal(t, rune(0), delimiter)
}

func TestTablo_DetectFieldDelimiter_MismatchedFieldCountsReturnZero(t *testing.T) {
	tbl := &Tablo{}

	delimiter := tbl.detectFieldDelimiter([]string{
		"name,age",
		"vigo,42,istanbul",
	})

	assert.Equal(t, rune(0), delimiter)
}

func TestTablo_BuildJSONDataset_AutoDetectsCSVDelimiter(t *testing.T) {
	tbl := &Tablo{}

	dataset := tbl.buildJSONDataset([]string{
		`name,age`,
		`vigo,42`,
	})

	assert.True(t, dataset.hasHeader)
	assert.Equal(t, []string{"name", "age"}, dataset.headers)
	assert.Equal(t, [][]string{{"vigo", "42"}}, dataset.rows)
}

func TestTablo_BuildJSONDataset_FilterIndexesTakePrecedence(t *testing.T) {
	tbl := &Tablo{
		Args:           []string{"name"},
		FieldDelimiter: '|',
		FilterIndexes:  []int{1},
	}

	dataset := tbl.buildJSONDataset([]string{"name|age", "vigo|42"})

	assert.False(t, dataset.hasHeader)
	assert.Empty(t, dataset.headers)
	assert.Equal(t, [][]string{{"age"}, {"42"}}, dataset.rows)
}

func TestTablo_BuildJSONDataset_FilterIndexesWithNoHeaders_SkipsDetectedHeaderRow(t *testing.T) {
	tbl := &Tablo{
		FieldDelimiter: ';',
		FilterIndexes:  []int{0, 2},
		HideHeaders:    true,
	}

	dataset := tbl.buildJSONDataset([]string{
		"Username;Identifier;First name;Last name",
		"booker12;9012;Rachel;Booker",
		"grey07;2070;Laura;Grey",
	})

	assert.False(t, dataset.hasHeader)
	assert.Empty(t, dataset.headers)
	assert.Equal(t, [][]string{{"booker12", "Rachel"}, {"grey07", "Laura"}}, dataset.rows)
}

func TestTablo_ShouldSkipFirstRow_FilterIndexesTakePrecedence(t *testing.T) {
	tbl := &Tablo{
		Args:          []string{"name"},
		FilterIndexes: []int{1},
	}

	skip := tbl.shouldSkipFirstRow([]string{"name|age", "vigo|42"})

	assert.False(t, skip)
}

func TestTablo_ShouldSkipFirstRow_NoHeadersInSmartMode_DoesNotSkip(t *testing.T) {
	tbl := &Tablo{
		HideHeaders: true,
	}

	skip := tbl.shouldSkipFirstRow([]string{"name  age", "vigo  42"})

	assert.False(t, skip)
}

func TestTablo_RenderJSON_WithoutHeaders_ReturnsWriteError(t *testing.T) {
	writeErr := errors.New("write failed")
	tbl := &Tablo{
		Output:         &errorWriteCloser{err: writeErr},
		FieldDelimiter: '|',
		FilterIndexes:  []int{1},
	}

	err := tbl.renderJSON([]string{"hello|world"})

	assert.ErrorIs(t, err, writeErr)
}

func TestTablo_RenderJSON_WithHeaders_ReturnsWriteError(t *testing.T) {
	writeErr := errors.New("write failed")
	tbl := &Tablo{
		Output:         &errorWriteCloser{err: writeErr},
		FieldDelimiter: '|',
	}

	err := tbl.renderJSON([]string{"name|age", "vigo|42"})

	assert.ErrorIs(t, err, writeErr)
}
