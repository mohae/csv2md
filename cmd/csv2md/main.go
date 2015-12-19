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
	flag.BoolVar(&format, "format", false, "use format file; location inferred from input")
	flag.BoolVar(&format, "f", false, "short flag for -format")
	flag.StringVar(&formatFile, "formatfile", "", "path to the format file; mutually exclusive with -format")
	flag.StringVar(&formatFile, "m", "", "short flag for -formatfile")
	flag.StringVar(&input, "input", "stdin", "input source")
	flag.StringVar(&input, "i", "stdin", "short flag for -input")
	flag.BoolVar(&lazyQuotes, "lazyquotes", false, "allow lazy quotes")
	flag.BoolVar(&lazyQuotes, "l", false, "short flag for -lazyquotes")
	flag.StringVar(&newLine, "newline", "\n", "newline sequence")
	flag.StringVar(&newLine, "n", "\n", "short flag for -newline")
	flag.BoolVar(&noHeaderRecord, "noheaderrecord", false, "CSV data does not include a header record")
	flag.BoolVar(&noHeaderRecord, "r", false, "short flag for -noheaderrecord")
	flag.StringVar(&output, "output", "stdout", "output destination")
	flag.StringVar(&output, "o", "stdout", "short flag for -output")
	flag.StringVar(&separator, "separator", ",", "field separator")
	flag.StringVar(&separator, "s", ",", "short flag for -s")
	flag.BoolVar(&trimLeadingSpace, "trimleadingspace", false, "trim leading space")
	flag.BoolVar(&trimLeadingSpace, "t", false, "short flag for -trimleadingspace")
	flag.BoolVar(&help, "help", false, "csv2md help")
	flag.BoolVar(&help, "h", false, "short flag for -help")
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
	if flag.NArg() > 0 {
		flag.Usage()
		return 2
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
		flag.Usage()
		return 0
	}
	var in, out, formatR *os.File
	var err error
	// set input
	in = os.Stdin
	if input != "stdin" {
		in, err = os.Open(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "input file error: %s\n", err)
			return 1
		}
	}
	defer in.Close()
	// set output
	out = os.Stdout
	if output != "stdout" {
		out, err = os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "output file error: %s\n", err)
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
			fmt.Fprintf(os.Stderr, "format file error: %s\n", err)
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
		fmt.Fprintf(os.Stderr, "transmogrifierication error: %s\n", err)
		return 1
	}
	return 0
}
