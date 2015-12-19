csv2md CLI
==========

csv2md is a CLI program that converts CSV-encoded data into a GitHub Flavored Markdown table.

The input can either be piped in from stdin or specified using either the `-i` or `-input` flag.  The output defaults to stdout, or can be specified using either the `-o` or `-output` flag.  If the CSV data does not include a field name record, the field names can be specified in a format file.  When a format file is used, the field names defined in the file will be used even if the input data contains a header record.  The format file can also be used to define field formatting.

## Format file
A format file can be defined for the CSV-encoded data.  Format files are CSV-encoded.  Format files can define field names, field alignment, and field styling.  A format file consists of up to 3 rows.

The first row of the format file contains the field names to be used as the table column names in the generated Markdown.  If a field value is empty, the CSV data's header record value for that field will be used instead, if the CSV data has a header record.

The second row of the format file, if it exists, contains the field alignment information.  Any field in this row that does not have a value will be unjustified in the resulting Markdown table.  This row is optional, unless the text styling is also being defined.  Valid values:

    Justification|Valid Values  
    :--|:--  
    Left|l, left, :--  
    Centered|c, center, centered, :--:  
    Right|r, right, --:  


The third row of the format file, if it exists, contains the text styling information for fields.  Any field in this row that does not have a value will not have styling applied in the resulting Markdown table.  This row is optional.  Valid values:  

    Styling|Valid Values
    :--|:--
    __Bold__|b, bold, __  
    _Italic_|i, italic, italics, _  
    ~~Strikethrough~~|s, strikethrough, ~~  

### format flag

The `-format`, or `-f`, flag is a bool flag that lets the program know if there is a format file for the data.  This flag can only be used when either the `-i` or `-input` flag is used.  csv2md will infer the format file name by replacing the specified input file extension with `.fmt`; e.g. `path/to/data.csv`'s format file would be `path/to/data.fmt`.  If the file cannot be found, an error will occur.  If the format file location needs to be specified, either the `-formatfile` or `-m` flag should be used instead.

### formatfile flag

The `-formatfile`, or `-m`, flag is a string flag that allows you to specify the location of the format file that should be used when creating the table Markdown.  If this file does not exist, an error will occur.

## Flags

Flag|Short|Default|Description  
:--|:--:|:--|:--  
format|f|false|use format file; location inferred from input  
formatfile|m||path to the format file; mutually exclusive with -format  
input|i|stding|input source
lazyquotes|l|false|allow lazy quotes  
newline|n|\n|newline sequence  
noheaderrecord|r|false|CSV data does not include a header record  
output|o|stdout|output destination  
separator|s|,|field separator  
trimleadingspace|t|false|trim leading space  
help|h|false|csv2md help  
