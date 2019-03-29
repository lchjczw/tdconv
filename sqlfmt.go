package tdconv

import (
	"fmt"
	"io"
	"strings"
)

// SQLFormatter is a formatter to output the table definision as SQL.
type SQLFormatter struct {
	header string
}

// NewSQLFormatter creates a new SQLFormatter.
// You can change some parameters of the SQLFormatter with SQLFormatOption.
func NewSQLFormatter(options ...SQLFormatOption) (*SQLFormatter, error) {
	sf := SQLFormatter{
		header: "# SQL generated by tdconv. DO NOT EDIT.\n# See more details at https://github.com/takuoki/tdconv\n",
	}
	for _, opt := range options {
		err := opt(&sf)
		if err != nil {
			return nil, err
		}
	}
	return &sf, nil
}

// SQLFormatOption changes some parameters of the SQLFormatter.
type SQLFormatOption func(*SQLFormatter) error

// SQLHeader changes the header string.
func SQLHeader(header string) SQLFormatOption {
	return func(sf *SQLFormatter) error {
		sf.header = header
		return nil
	}
}

// Fprint outputs the table definision as SQL.
func (f *SQLFormatter) Fprint(w io.Writer, t *Table) {

	if f == nil || t == nil {
		return
	}

	fmt.Fprintf(w, "%[1]sDROP TABLE IF EXISTS %[2]s;\nCREATE TABLE `%[2]s` (\n", f.header, t.Name)

	for _, c := range t.Columns {
		es := make([]string, 0, 6)
		es = append(es, "    `"+c.Name+"`")
		es = append(es, c.Type)
		if c.NotNull {
			es = append(es, "NOT NULL")
		}
		if c.Option != "" {
			es = append(es, c.Option)
		}
		if c.Unique {
			es = append(es, "UNIQUE")
		}
		if c.Comment != "" {
			es = append(es, "COMMENT '"+c.Comment+"'")
		}
		fmt.Fprintf(w, "%s,\n", strings.Join(es, " "))
	}

	fmt.Fprintf(w, "    PRIMARY KEY (%s)", strings.Join(t.PKeyColumns, ", "))

	for _, k := range t.UniqueKeys {
		fmt.Fprintf(w, ",\n    UNIQUE KEY `%s` (%s)", k.Name, strings.Join(k.Columns, ", "))
	}
	for _, k := range t.IndexKeys {
		fmt.Fprintf(w, ",\n    UNIQUE KEY `%s` (%s)", k.Name, strings.Join(k.Columns, ", "))
	}

	fmt.Fprintf(w, "\n);")
}
