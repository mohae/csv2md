# csv2md
[![Build Status](https://travis-ci.org/mohae/csv2md.png)](https://travis-ci.org/mohae/csv2md)

csv2md transmogrifies CSV-encoded data into GitHub Flavored Markdown, GFM, tables.

Formatting of fields is supported. Text can have no justification, or be: left justified, centered, or right justified.  Text can either be un-styled or styled with bold, italic, or strikethrough styling.  Formatting is per column, field, and does not apply to the table header row, record.

For more details see https://help.github.com/articles/github-flavored-markdown/#tables.

An example implementation and cli app can be found at https://github.com/mohae/csv2md/tree/master/cmd/csv2md.  Documentation on usage of the CLI app is in the [cli's README](https://github.com/mohae/csv2md/tree/master/cmd/csv2md/readme)

## Docs
https://godoc.org/github.com/mohae/csv2md
