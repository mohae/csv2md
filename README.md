# csv2md
csv2md transmogrifies CSV-encoded data intto GitHub Flavored Mardwon, GFM, tables.

Formatting of fields is supported. Text can be unjustified, left justified, centered, or right justified.  Text can also either be unstyled or styled with bold, italic, or strikethrough.  Formatting is per column and does not apply to the table header row.

For more details see https://help.github.com/articles/github-flavored-markdown/#tables.

An example implementation and cli app can be found at https://github.com/mohae/csv2md/tree/master/cmd/csv2md.  Documentation on usage of the CLI app is in the [cli's README](https://github.com/mohae/csv2md/blob/master/cmd/csv2md/README.md)
