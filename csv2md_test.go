package csv2md

import (
	"bytes"
	"testing"
)

func TestSetFieldNames(t *testing.T) {
	tests := []struct {
		fields []string
	}{
		{[]string{""}},
		{[]string{"a", "b"}},
		{[]string{"a", "b", "c", "d"}},
	}
	for i, test := range tests {
		calvin := Transmogrifier{}
		calvin.SetFieldNames(test.fields)
		if len(calvin.fieldNames) != len(test.fields) {
			t.Errorf("%d field names count: got %d want %d", i, len(calvin.fieldNames), len(test.fields))
		}
		for j, v := range calvin.fieldNames {
			if v != test.fields[j] {
				t.Errorf("%d,%d: got %s want %s", i, j, v, test.fields[j])
			}
		}
	}
}

func TestSetFieldAlignment(t *testing.T) {
	tests := []struct {
		fields   []string
		expected []string
	}{
		{},
		{[]string{""}, []string{"---"}},
		{[]string{"", ":--", "l", "", ":--:", "", "--:", "-"}, []string{"---", ":--", ":--", "---", ":--:", "---", "--:", "---"}},
		{[]string{"l", "left", "r", "right", "c", "center", "centered", ""}, []string{":--", ":--", "--:", "--:", ":--:", ":--:", ":--:", "---"}},
	}
	for i, test := range tests {
		calvin := Transmogrifier{}
		calvin.SetFieldAlignment(test.fields)
		if len(calvin.fieldAlignment) != len(test.expected) {
			t.Errorf("%d field alignment count: got %d want %d", i, len(calvin.fieldAlignment), len(test.expected))
		}
		for j, v := range calvin.fieldAlignment {
			if v != test.expected[j] {
				t.Errorf("%d, %d: got %s want %s", i, j, v, test.expected[j])
			}
		}
	}
}

func TestSetFieldStyle(t *testing.T) {
	tests := []struct {
		fields   []string
		expected []string
	}{
		{},
		{[]string{""}, []string{""}},
		{[]string{"", "_", "italic", "", "__", "~~", "adsf"}, []string{"", "_", "_", "", "__", "~~", ""}},
		{[]string{"i", "italic", "italics", "b", "bold", "s", "strikethrough", "z", ""}, []string{"_", "_", "_", "__", "__", "~~", "~~", "", ""}},
	}
	for i, test := range tests {
		calvin := Transmogrifier{}
		calvin.SetFieldStyle(test.fields)
		if len(calvin.fieldStyle) != len(test.expected) {
			t.Errorf("%d field style count: got %d want %d", i, len(calvin.fieldStyle), len(test.expected))
		}
		for j, v := range calvin.fieldStyle {
			if v != test.expected[j] {
				t.Errorf("%d, %d: got %s want %s", i, j, v, test.expected[j])
			}
		}
	}
}

func TestSetFmt(t *testing.T) {
	var b []byte
	var w bytes.Buffer
	r := bytes.NewReader(b)
	h := []byte("a,b,c,d\nl,c,r,\ni,b,s,\n")
	hR := bytes.NewReader(h)
	expectedNames := []string{"a", "b", "c", "d"}
	expectedAlignment := []string{":--", ":--:", "--:", "---"}
	expectedFmt := []string{"_", "__", "~~", ""}
	calvin := NewTransmogrifier(r, &w)
	err := calvin.SetFmt(hR)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	if len(calvin.fieldNames) != len(expectedNames) {
		t.Errorf("fieldNames len was %d, want %d", len(calvin.fieldNames), len(expectedNames))
	} else {
		for i, v := range calvin.fieldNames {
			if v != expectedNames[i] {
				t.Errorf("%d: fieldName was %s, want %s", i, v, expectedNames[i])
			}
		}
	}
	if len(calvin.fieldAlignment) != len(expectedAlignment) {
		t.Errorf("fieldAlignment len was %d, want %d", len(calvin.fieldAlignment), len(expectedAlignment))
	} else {
		for i, v := range calvin.fieldAlignment {
			if v != expectedAlignment[i] {
				t.Errorf("%d: fieldAlignment was %s, want %s", i, v, expectedAlignment[i])
			}
		}
	}
	if len(calvin.fieldStyle) != len(expectedFmt) {
		t.Errorf("fieldStyle len was %d, want %d", len(calvin.fieldStyle), len(expectedFmt))
	} else {
		for i, v := range calvin.fieldStyle {
			if v != expectedFmt[i] {
				t.Errorf("%d: fieldStyle was %s, want %s", i, v, expectedFmt[i])
			}
		}
	}
}

func TestMDTable(t *testing.T) {
	csvData := []byte("Manufacturer,Model,Type,Year\nFord,Focus,Sedan,2015\nChevy,Malibu,Sedan,2015\n")
	format := []byte("Make,Model,Type,Yr\nc, l, left, right\nbold, italic, ,strikethrough\n")
	tests := []struct {
		useFmt    bool
		hasHeader bool
		data      []byte
		expected  string
	}{
		{false, true, csvData, "Manufacturer|Model|Type|Year  \n---|---|---|---  \nFord|Focus|Sedan|2015  \nChevy|Malibu|Sedan|2015  \n"},
		{true, false, csvData, "Make|Model|Type|Yr  \n:--:|:--|:--|--:  \n__Manufacturer__|_Model_|Type|~~Year~~  \n__Ford__|_Focus_|Sedan|~~2015~~  \n__Chevy__|_Malibu_|Sedan|~~2015~~  \n"},
		{true, true, csvData, "Make|Model|Type|Yr  \n:--:|:--|:--|--:  \n__Ford__|_Focus_|Sedan|~~2015~~  \n__Chevy__|_Malibu_|Sedan|~~2015~~  \n"},
		{false, false, []byte("Manufacturer,Model,Type,Year\n,Focus,Sedan,2015\n,Malibu,Sedan,2015\n"),
			"Manufacturer|Model|Type|Year  \n |Focus|Sedan|2015  \n |Malibu|Sedan|2015  \n"},
	}
	for i, test := range tests {
		var w bytes.Buffer
		r := bytes.NewReader(test.data)
		calvin := NewTransmogrifier(r, &w)
		calvin.HasHeaderRecord = test.hasHeader
		if test.useFmt {
			fR := bytes.NewReader(format)
			err := calvin.SetFmt(fR)
			if err != nil {
				t.Errorf("%d: unexpected error setting format: %s", i, err)
				continue
			}
		}
		err := calvin.MDTable()
		if err != nil {
			t.Errorf("%d: unexpected error creating mdtable: %s", i, err)
			continue
		}
		if w.String() != test.expected {
			t.Errorf("%d: got %q want %q", i, w.String(), test.expected)
		}
	}
}
