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
	"fmt"
	"io"
	"strings"
)

const (
	left  = ":--"
	centered = ":--:"
	right = "--:"
	none = "---"
	italic = "_"
	bold = "__"
	strikethrough = "~~"

)
type ShortWriteError struct {
	n         int
	written   int
	operation string
}

func (e ShortWriteError) Error() string {
	return fmt.Sprintf("short write of %s: wrote %d bytes of %d", e.operation, e.n, e.written)
}

// Transmogrifier turns CSV data into a markdown table
type Transmogrifier struct {
	// HasHeaderRecord specifies whether or not the CSV-encoded data's
	// first record has field names.  If false, either the field names
	// must be set; either by calling SetFieldNames or SetFmt.  In either
	// case, the number of fields must match the number of fields per
	// record in the CSV data.
	HasHeaderRecord bool
	// CSV is a csv.Reader.  The fields are exposed so that the caller
	// can configure.
	CSV *csv.Reader
	w io.Writer
	fieldNames []string
	fieldAlignment []string
	fieldFmt []string
	newLine string
	rBytes int
	wBytes int
}

func NewTransmogrifier(r io.Reader, w io.Writer) *Transmogrifier {
	return &Transmogrifier{HasHeaderRecord: true, CSV: csv.NewReader(r), w: w, newLine: "  \n"}
}

// SetNewLine sets the newLine value based on the received value.  If the
// received value is not recognized, nothing is done.
//
// Recognized values:
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

func (t *Transmogrifier) NewLine() string {
	return t.newLine
}

func (t *Transmogrifier) SetFieldNames(vals []string) {
	t.fieldNames = append(t.fieldNames, vals...)
	return
}

func (t *Transmogrifier) SetFieldAlignment(vals []string) {
	for _, v := range vals {
		v = strings.TrimSpace(strings.ToLower(v))
		switch v {
		case "l", "left":
			t.fieldAlignment = append(t.fieldAlignment, left)
		case "c", "center", "centered":
			t.fieldAlignment = append(t.fieldAlignment, centered)
		case "r", "right":
			t.fieldAlignment = append(t.fieldAlignment, right)
		default:
			t.fieldAlignment = append(t.fieldAlignment, none)
		}
	}
	return
}

func (t *Transmogrifier) SetFieldFmt(vals []string)  {
	for _, v := range vals {
		v = strings.TrimSpace(strings.ToLower(v))
		switch v {
		case "b", "bold":
			t.fieldFmt = append(t.fieldFmt, bold)
		case "i", "italic", "italics":
			t.fieldFmt = append(t.fieldFmt, italic)
		case "s", "strikethrough":
			t.fieldFmt = append(t.fieldFmt, strikethrough)
		default:
			t.fieldFmt = append(t.fieldFmt, "")
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
			return fmt.Errorf("no format data found")
		}
		// first row is assumed to be the field names
		t.fieldNames = make([]string, len(records[0]))
		copy(t.fieldNames, records[0])
		// second row is field alignment, if it exists
		if len(records) > 1 {
			t.SetFieldAlignment(records[1])
		}
		// third row is text formatting for each field, if it exists
		if len(records) > 2 {
			t.SetFieldFmt(records[2])
		}
		return nil
}

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
	end := len(fields) - 1
	for i, field := range fields {
		if i < end {
			field = fmt.Sprintf("%s|", field)
		}
		t.wBytes, err = t.w.Write([]byte(field))
		if err != nil {
			return err
		}
	}
	t.wBytes, err = t.w.Write([]byte(t.newLine))
	if err != nil {
		return err
	}
	// write the header record separator
	if len(t.fieldAlignment) == 0 {
		// no field alignment was set, just write out the separator row
		for i := 0; i < len(fields); i++ {
			val := none
			if i < end {
				val = fmt.Sprintf("%s|", val)
			}
			t.wBytes, err = t.w.Write([]byte(val))
			if err != nil {
				return err
			}
		}
		t.wBytes, err = t.w.Write([]byte(t.newLine))
		return nil
	}
	end = len(t.fieldAlignment) - 1
	for i, field := range t.fieldAlignment {
		if i < end {
			field = fmt.Sprintf("%s|", field)
		}
		t.wBytes, err = t.w.Write([]byte(field))
	}
	t.wBytes, err = t.w.Write([]byte(t.newLine))
	return err
}

func (t *Transmogrifier) writeRecord(fields []string) error {
	var err error
	format := len(t.fieldFmt) > 0
	end := len(fields) - 1
	for i, field := range fields {
		if format {
			field = fmt.Sprintf("%s%s%s", t.fieldFmt[i], field, t.fieldFmt[i])
		}
		if i < end  {
			field = fmt.Sprintf("%s|", field)
		}
		t.wBytes, err = t.w.Write([]byte(field))
		if err != nil {
			return err
		}
	}
	t.wBytes, err = t.w.Write([]byte(t.newLine))
	return err
}
