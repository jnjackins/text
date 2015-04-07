// package column implements an io.Writer which formats
// input lines into columns.
package column // import "sigint.ca/text/column"

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// A Writer is an io.Writer which filters text by arranging it into columns.
type Writer struct {
	buf      *bytes.Buffer
	w        io.Writer
	maxwidth int
	colwidth int
}

// NewWriter returns a new column.Writer. Text written to this writer will be
// arranged so that its combined width does not exceed the given width, and then
// written to w when flushed by calling Flush().
func NewWriter(w io.Writer, width int) *Writer {
	return &Writer{
		buf:      &bytes.Buffer{},
		w:        w,
		maxwidth: width,
	}
}

// Write writes p to an internal buffer. No writes are done to the backing io.Writer
// until Flush is called.
func (w *Writer) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}

type column struct {
	words []string
}

// Flush performs the columnation and writes the results to the column.Writer's
// backing io.Writer.
func (w *Writer) Flush() error {
	words := strings.Split(w.buf.String(), "\n")
	w.colwidth = maxlen(words)
	cols := make([]column, 1)
	cols[0].words = words
	for w.split(words, &cols) {
	}
	return w.print(cols)
}

// maxlen returns the maximum length, in runes, of the strings in words
func maxlen(words []string) int {
	var max int
	for i := range words {
		l := len([]rune(words[i]))
		if l > max {
			max = l
		}
	}
	return max
}

// split returns true if the split was successful, or false if cols is already
// maximally columnated.
func (w *Writer) split(words []string, cols *[]column) bool {
	// try to become one column wider
	newcols := make([]column, len(*cols)+1)
	percol := len(words) / len(newcols)
	if len(words)%len(newcols) != 0 {
		percol++
	}
	for colnum := range newcols {
		i, j := percol*colnum, percol*colnum+percol
		if j > len(words) {
			j = len(words)
		}

		// empty columns are possible, bail out if we've reached one.
		// otherwise, slice out some words for the column.
		if i < len(words) {
			colwords := words[i:j]
			newcols[colnum] = column{words: colwords}
		} else {
			break
		}
	}

	// if newcols is too wide, discard it and stop
	if w.totalwidth(newcols) >= w.maxwidth {
		return false
	}

	// otherwise, tell the caller to continue splitting
	*cols = newcols
	return true
}

// totalwidth returns the total width of cols.
func (w *Writer) totalwidth(cols []column) int {
	width := (w.colwidth + 1) * (len(cols) - 1)
	var lastwidth int
	for _, word := range cols[len(cols)-1].words {
		if len(word) > lastwidth {
			lastwidth = len(word)
		}
	}
	return width + lastwidth
}

// print writes the columns to the backing io.Writer.
func (w *Writer) print(cols []column) error {
	rowc := len(cols[0].words)
	for i := 0; i < rowc; i++ {
		for j := range cols {
			if i >= len(cols[j].words) {
				break // done this row
			}
			if j < len(cols)-1 {
				_, err := fmt.Fprintf(w.w, "%-*s", w.colwidth+1, cols[j].words[i])
				if err != nil {
					return err
				}
			} else {
				_, err := fmt.Fprintf(w.w, "%s", cols[j].words[i])
				if err != nil {
					return err
				}
			}
		}
		_, err := fmt.Fprintln(w.w)
		if err != nil {
			return err
		}
	}
	return nil
}
