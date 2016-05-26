// Package csv2md transforms CSV-encoded data into Github Flavored
// Mardown, GFM, tables.  The field names, field alignment, and field
// text effects can be specified: either by providing a format file or
// by configuring the Transmogrifier.  If the field names have not been
// set, the first record of the CSV-encoded data should have a header
// record.
//
// It is assumed that all CSV-encoded data contains a header record,
// even if the field names have been set,  unless the HasHeaderRecord is
// set to false.  If the field names has been set and the CSV-encoded
// data has a header record, the first record in the data will be ignored.
package csv2md

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	left          = ":--"
	centered      = ":--:"
	right         = "--:"
	none          = "---"
	italic        = "_"
	bold          = "__"
	strikethrough = "~~"
)

// ShortWriteError occurs when the number of bytes written is less than
// the number of bytes to be written.
type ShortWriteError struct {
	n         int
	written   int
	operation string
}

func (e ShortWriteError) Error() string {
	return fmt.Sprintf("%s: short write of, wrote %d of %d bytes", e.operation, e.n, e.written)
}

// ErrNoFormatData occurs when no data is found in the provided reader.
var ErrNoFormatData = errors.New("no format data")

// Transmogrifier turns CSV data into a markdown table
type Transmogrifier struct {
	// HasHeaderRecord specifies whether or not the CSV-encoded data's
	// first record has field names.  If false, either the field names
	// must be set; either by calling SetFieldNames or SetFmt.  In either
	// case, the number of fields must match the number of fields per
	// record in the CSV data.
	HasHeaderRecord bool
	// CSV is a csv.Reader.  This is exported so that the caller can
	// can configure the CSV reader.
	CSV            *csv.Reader
	w              io.Writer
	fieldNames     []string
	fieldAlignment []string
	fieldStyle     []string
	newLine        string
	rBytes         int64
	wBytes         int64
}

// NewTransmogrifier returns an initialized Transmogrifier for
// transmogrifierication of CSV-encoded data to GitHub Flavored Markdown
// tables.
func NewTransmogrifier(r io.Reader, w io.Writer) *Transmogrifier {
	return &Transmogrifier{HasHeaderRecord: true, CSV: csv.NewReader(r), w: w, newLine: "  \n"}
}

// BytesWritten returns the number of bytes written to the writer.
func (t *Transmogrifier) BytesWritten() int64 {
	return t.wBytes
}

// SetNewLine sets the newLine value based on the received value.  If the
// received value is not recognized, nothing is done.
//
// Valid new line values:
//    * Carriage Return (new line)
//      * cr
//      * CR
//      * \n
//    * Line Feed
//      * lf
//      * LF
//      * \r
//    * Carriage Return/Line Feed
//      * crlf
//      * CRLF
//      * \r\n
//
// For a new line to occur, Markdown requires the line to terminate with
// either two spaces, "  ", or have a double line feed.  The newLine is
// prefixed with two spaces.
func (t *Transmogrifier) SetNewLine(s string) {
	switch s {
	case "cr", "CR", "\n":
		t.newLine = "  \n"
	case "lf", "LF", "\r":
		t.newLine = "  \r"
	case "crlf", "CRLF", "\r\n":
		t.newLine = "   \r"
	}
}

// NewLine returns the current new line sequence.
func (t *Transmogrifier) NewLine() string {
	return t.newLine
}

// SetFieldNames sets the names for each field; these values are used in
// the table header as each column's, field's, name.
func (t *Transmogrifier) SetFieldNames(vals []string) {
	t.fieldNames = append(t.fieldNames, vals...)
	return
}

// SetFieldAlignment sets the alignment, justification, used for each
// field in the table.
//
// Valid alignment values:
//   * Left justification
//     * l
//     * left
//     * :--
//   * Centered text
//     * c
//     * centered
//     * center
//     * :--:
//   * Right justification
//     * r
//     * right
//     * --:
//   * No justification
//     * empty string
func (t *Transmogrifier) SetFieldAlignment(vals []string) {
	for _, v := range vals {
		v = strings.TrimSpace(strings.ToLower(v))
		switch v {
		case "l", "left", left:
			t.fieldAlignment = append(t.fieldAlignment, left)
		case "c", "center", "centered", centered:
			t.fieldAlignment = append(t.fieldAlignment, centered)
		case "r", "right", right:
			t.fieldAlignment = append(t.fieldAlignment, right)
		default:
			t.fieldAlignment = append(t.fieldAlignment, none)
		}
	}
	return
}

// SetFieldStyle sets the text styling for a record's field.
// Accepted values:
//    * Bold
//      * b
//      * bold
//      * __
//    * Italics
//      * i
//      * italics
//      * italic
//      * _
//    * Strikethrough
//      * s
//      * strikethrough
//      * ~~
//    * No text styling
//      * empty string
func (t *Transmogrifier) SetFieldStyle(vals []string) {
	for _, v := range vals {
		v = strings.TrimSpace(strings.ToLower(v))
		switch v {
		case "b", "bold", bold:
			t.fieldStyle = append(t.fieldStyle, bold)
		case "i", "italic", "italics", italic:
			t.fieldStyle = append(t.fieldStyle, italic)
		case "s", "strikethrough", strikethrough:
			t.fieldStyle = append(t.fieldStyle, strikethrough)
		default:
			t.fieldStyle = append(t.fieldStyle, "")
		}
	}
}

// SetFmt takes a reader and reads the format information from it as CSV
// encoded data.  The CSV reader used to read the format information is
// configured to be consistent with CSV's configuration under the assumption
// that the CSV data in the format file will be encoded the same way as the
// actual CSV data; e.g. if the CSV data is tab delimited, the format file
// will also be tab delimited.
func (t *Transmogrifier) SetFmt(r io.Reader) error {
	c := csv.NewReader(r)
	// make sure this reader's settings are consistent with CSV's
	c.Comma = t.CSV.Comma
	c.Comment = t.CSV.Comment
	c.FieldsPerRecord = t.CSV.FieldsPerRecord
	c.LazyQuotes = t.CSV.LazyQuotes
	c.TrailingComma = t.CSV.TrailingComma
	c.TrimLeadingSpace = t.CSV.TrimLeadingSpace
	records, err := c.ReadAll()
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return ErrNoFormatData
	}
	// first row is assumed to be the field names
	t.fieldNames = make([]string, len(records[0]))
	copy(t.fieldNames, records[0])
	// second row is field alignment, if it exists
	if len(records) > 1 {
		t.SetFieldAlignment(records[1])
	}
	// third row is text styling for each field, if it exists
	if len(records) > 2 {
		t.SetFieldStyle(records[2])
	}
	return nil
}

// MDTable reads from the configured reader, CSV, transforms the data into
// a GitHub Flavored Markdown table, applying justification and text
// styling, and writes the resulting bytes to the Transmogrifier's writer.
func (t *Transmogrifier) MDTable() error {
	// if the field names are set, write those first
	if len(t.fieldNames) > 0 {
		err := t.writeHeaderRecord(t.fieldNames)
		if err != nil {
			return err
		}
	}
	// read until EOF
	var row int
	for {
		row++
		record, err := t.CSV.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if row == 1 && t.HasHeaderRecord {
			if len(t.fieldNames) > 0 {
				continue
			}
			err = t.writeHeaderRecord(record)
			if err != nil {
				return err
			}
			continue
		}
		err = t.writeRecord(record)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Transmogrifier) writeHeaderRecord(fields []string) error {
	var err error
	var n int
	end := len(fields) - 1
	for i, field := range fields {
		if i < end {
			field = fmt.Sprintf("%s|", field)
		}
		n, err = t.w.Write([]byte(field))
		if err != nil {
			return err
		}
		if n != len(field) {
			return ShortWriteError{n: len(field), written: n, operation: "header field"}
		}
		t.wBytes += int64(n)
	}
	n, err = t.w.Write([]byte(t.newLine))
	if err != nil {
		return err
	}
	if n != len(t.newLine) {
		return ShortWriteError{n: len(t.newLine), written: n, operation: "new line"}
	}
	t.wBytes += int64(n)
	// write the header record separator
	if len(t.fieldAlignment) == 0 {
		// no field alignment was set, just write out the separator row
		for i := 0; i < len(fields); i++ {
			val := none
			if i < end {
				val = fmt.Sprintf("%s|", val)
			}
			n, err = t.w.Write([]byte(val))
			if err != nil {
				return err
			}
			if n != len(val) {
				return ShortWriteError{n: len(val), written: n, operation: "header row separator"}
			}
			t.wBytes += int64(n)
		}
		n, err = t.w.Write([]byte(t.newLine))
		if err != nil {
			return err
		}
		if n != len(t.newLine) {
			return ShortWriteError{n: len(t.newLine), written: n, operation: "new line"}
		}
		t.wBytes += int64(n)
		return nil
	}
	end = len(t.fieldAlignment) - 1
	for i, field := range t.fieldAlignment {
		if i < end {
			field = fmt.Sprintf("%s|", field)
		}
		n, err = t.w.Write([]byte(field))
		if err != nil {
			return err
		}
		if n != len(field) {
			return ShortWriteError{n: len(field), written: n, operation: "header row separator"}
		}
		t.wBytes += int64(n)
	}
	n, err = t.w.Write([]byte(t.newLine))
	if err != nil {
		return err
	}
	if n != len(t.newLine) {
		return ShortWriteError{n: len(t.newLine), written: n, operation: "new line"}
	}
	t.wBytes += int64(n)
	return err
}

func (t *Transmogrifier) writeRecord(fields []string) error {
	var err error
	var n int
	format := len(t.fieldStyle) > 0
	end := len(fields) - 1
	for i, field := range fields {
		// if the field is empty, add a space to indicate to MD that there is a value
		// otherwise columns may not end up in the correct spot.
		if field == "" {
			field = " "
		}
		if format {
			field = fmt.Sprintf("%s%s%s", t.fieldStyle[i], field, t.fieldStyle[i])
		}
		if i < end {
			field = fmt.Sprintf("%s|", field)
		}
		n, err = t.w.Write([]byte(field))
		if err != nil {
			return err
		}
		if n != len(field) {
			return ShortWriteError{n: len(field), written: n, operation: "record field"}
		}
		t.wBytes += int64(n)
	}
	n, err = t.w.Write([]byte(t.newLine))
	if err != nil {
		return err
	}
	if n != len(t.newLine) {
		return ShortWriteError{n: len(t.newLine), written: n, operation: "new line"}
	}
	t.wBytes += int64(n)
	return err
}
