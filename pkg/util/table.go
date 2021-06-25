// RAINBOND, Application Management Platform
// Copyright (C) 2014-2017 Goodrain Co., Ltd.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Rainbond,
// one or multiple Commercial Licenses authorized by Goodrain Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package util

import (
	"bytes"
	"math"
	"regexp"
	"strings"
)

var ansiRegexp = regexp.MustCompile("\x1b[^m]*m")

// Table -
type Table struct {
	Rows    [][]string
	Columns [][]string
	Options *TableOptions

	HasHeader bool

	numColumns   int
	columnsWidth []int
}

// TableOptions -
type TableOptions struct {
	Padding      int
	UseSeparator bool
}

var defaultTableOptions = &TableOptions{
	Padding:      3,
	UseSeparator: false,
}

// NewTable -
func NewTable(rows [][]string, options *TableOptions) *Table {
	t := &Table{
		Options: options,
	}
	if t.Options == nil {
		t.Options = defaultTableOptions
	}
	if rows != nil {
		t.Rows = rows
		t.computeProperties()
	}
	return t
}

// SetHeader -
func (t *Table) SetHeader(header []string) {
	t.HasHeader = true
	// There is a better way to do this
	t.Rows = append([][]string{header}, t.Rows...)
	t.computeProperties()
}

// AddRow -
func (t *Table) AddRow(row []string) {
	t.Rows = append(t.Rows, row)
	t.computeProperties()
}

func (t *Table) computeProperties() {
	if len(t.Rows) > 0 {
		t.numColumns = len(t.Rows[0])
		t.columnsWidth = make([]int, t.numColumns)
		t.recalculate()
	}
}

func (t *Table) recalculate() {
	t.Columns = [][]string{}
	for i := 0; i < t.numColumns; i++ {
		t.Columns = append(t.Columns, []string{})
	}
	for _, row := range t.Rows {
		for j, cellContent := range row {
			t.Columns[j] = append(t.Columns[j], cellContent)
			t.columnsWidth[j] = int(math.Max(float64(visibleLen(cellContent)), float64(t.columnsWidth[j])))
		}
	}
}

// Render -
func (t *Table) Render() string {
	// allocate a 1k byte buffer
	bb := make([]byte, 0, 1024)
	buf := bytes.NewBuffer(bb)

	i := 0

	if t.HasHeader {
		if t.Options.UseSeparator {
			buf.WriteString(t.separatorLine())
			buf.WriteRune('\n')
		}
		for j := range t.Rows[0] {
			buf.WriteString(t.getCell(i, j))
		}
		i = 1
		buf.WriteRune('\n')
	}

	if t.Options.UseSeparator {
		buf.WriteString(t.separatorLine())
		buf.WriteRune('\n')
	}

	for i < len(t.Rows) {
		row := t.Rows[i]
		for j := range row {
			buf.WriteString(t.getCell(i, j))
		}
		if i < len(t.Rows)-1 {
			buf.WriteRune('\n')
		}
		i++
	}

	if t.Options.UseSeparator {
		buf.WriteRune('\n')
		buf.WriteString(t.separatorLine())
	}

	return buf.String()
}

func (t *Table) separatorLine() string {
	sep := "+"
	for _, w := range t.columnsWidth {
		sep += strings.Repeat("-", w+2*t.Options.Padding)
		sep += "+"
	}
	return sep
}

func (t *Table) getCell(row, col int) string {
	cellContent := t.Rows[row][col]
	spacePadding := strings.Repeat(" ", t.Options.Padding)

	var cellStr string

	if t.Options.UseSeparator {
		cellStr += "|"
		cellStr += spacePadding
	}

	cellStr += cellContent
	cellStr += strings.Repeat(" ", t.columnsWidth[col]-visibleLen(cellContent))
	cellStr += spacePadding

	if t.Options.UseSeparator {
		if col == t.numColumns-1 {
			cellStr += "|"
		}
	}

	return cellStr
}

func visibleLen(s string) int {
	return len(ansiRegexp.ReplaceAllLiteralString(s, ""))
}
