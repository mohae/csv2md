package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mohae/csv2md"
)

// flags
var (
	format           bool
	formatFile       string
	input            string
	help             bool
	lazyQuotes       bool
	newLine          string
	noHeaderRecord   bool
	output           string
	separator        string
	trimLeadingSpace bool
)

var prog = filepath.Base(os.Args[0])

func init() {
	flag.BoolVar(&format, "format", false, "if a format file should be used; this flag assumes that it is in the same location as input with a .fmt ext")
	flag.BoolVar(&format, "f", false, "the short flag for -format")
	flag.StringVar(&formatFile, "formatfile", "", "the path to the format file, if one, the format bool does not need to be specified when this flag is used")
	flag.StringVar(&formatFile, "m", "", "the short flag for -formatfile")
	flag.StringVar(&input, "input", "stdin", "the path to the input file; if not specified stdin is used")
	flag.StringVar(&input, "i", "stdin", "the short flag for -input")
	flag.BoolVar(&lazyQuotes, "lazyquotes", false, "allow lazy quotes")
	flag.BoolVar(&lazyQuotes, "l", false, "the short flag for -lazyquotes")
	flag.StringVar(&newLine, "newline", "\n", "the newline sequence to use")
	flag.StringVar(&newLine, "n", "\n", "the short flag for -newline")
	flag.BoolVar(&noHeaderRecord, "noheaderrecord", false, "if the CSV data includes field names; if false a format file should be used")
	flag.BoolVar(&noHeaderRecord, "r", false, "the short flag for -noheaderrecord")
	flag.StringVar(&output, "output", "stdout", "path to the output file; if not specified stdout is used")
	flag.StringVar(&output, "o", "stdout", "the short flag for -output")
	flag.StringVar(&separator, "separator", ",", "path to the output file; if not specified stdout is used")
	flag.StringVar(&separator, "s", ",", "the short flag for -output")
	flag.BoolVar(&trimLeadingSpace, "trimleadingspace", false, "trim leading space")
	flag.BoolVar(&trimLeadingSpace, "t", false, "the short flag for -trimleadingspace")
	flag.BoolVar(&help, "help", false, "csv2md help")
	flag.BoolVar(&help, "h", false, "the short flag for -help")
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s [OPTS]\n", prog)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Creates Github Style Markdown tables from CSV-encoded data\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

func main() {
	os.Exit(realMain())
}

func realMain() int {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	// check args; there shouldn't be any, but this is in case help was
	// used without the flag prefix
	args := flag.Args()
	for _, arg := range args {
		if arg == "help" {
			help = true
			break
		}
	}
	if help {
		usage()
		return 0
	}
	var in, out, formatR *os.File
	var err error
	// set input
	in = os.Stdin
	if input != "stdin" {
		in, err = os.Open(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "input file error: %s", err)
			return 1
		}
	}
	defer in.Close()
	// set output
	out = os.Stdout
	if output != "stdout" {
		out, err = os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "output file error: %s", err)
			return 1
		}
		defer out.Close()
	}

	// if formatting was specified but no format file was given, set the
	// format file to be the same as the input, replacing the input file's
	// extension with '.fnmt'
	if format && len(formatFile) == 0 {
		// if input is stdin error
		if input == "stdin" {
			fmt.Fprintln(os.Stderr, "cannot infer the format file location when using stdin for the input; when stdin is the input, the location must be specified using either the '-formatfile' or '-m' flag")
			return 1
		}
		// build the filepath from the input, if input is stdin error
		formatFile = fmt.Sprintf("%s.fmt", strings.TrimSuffix(input, filepath.Ext(input)))
	}
	// format stuff
	if len(formatFile) > 0 {
		// if the format file is specified use that
		formatR, err = os.OpenFile(formatFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "format file error: %s", err)
			return 1
		}
	}

	t := csv2md.NewTransmogrifier(in, out)
	if len(separator) > 0 {
		tmp := []rune(separator)
		t.CSV.Comma = tmp[0]
	}
	t.CSV.LazyQuotes = lazyQuotes
	t.CSV.TrimLeadingSpace = trimLeadingSpace
	t.SetNewLine(newLine)
	fmt.Printf("%q", t.NewLine())
	t.SetFmt(formatR)
	err = t.MDTable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "transmogrifierication error: %s", err)
		return 1
	}
	return 0
}
