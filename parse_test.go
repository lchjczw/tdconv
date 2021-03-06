package tdconv_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/takuoki/gostr"
	"github.com/takuoki/gsheets"
	"github.com/takuoki/tdconv"
)

var mustNewParser = func(options ...tdconv.ParseOption) *tdconv.Parser {
	p, err := tdconv.NewParser(options...)
	if err != nil {
		panic(err)
	}
	return p
}

func TestNewParser(t *testing.T) {

	cases := []struct {
		caseName string
		opts     []tdconv.ParseOption
		errMsg   string
	}{
		{
			caseName: "success: default",
		},
		{
			caseName: "failure: TableNamePos row",
			opts:     []tdconv.ParseOption{tdconv.TableNamePos(5, "C")},
			errMsg:   "Table name row must be smaller than the start row",
		},
		{
			caseName: "failure: TableNamePos column",
			opts:     []tdconv.ParseOption{tdconv.TableNamePos(1, "!")},
			errMsg:   "Unable to convert column string",
		},
		{
			caseName: "failure: StartRow",
			opts:     []tdconv.ParseOption{tdconv.StartRow(0)},
			errMsg:   "Start row must be greater than the table name row",
		},
		{
			caseName: "failure: TableNamePos -> StartRow",
			opts:     []tdconv.ParseOption{tdconv.TableNamePos(3, "C"), tdconv.StartRow(2)},
			errMsg:   "Start row must be greater than the table name row",
		},
		{
			caseName: "failure: StartRow -> TableNamePos",
			opts:     []tdconv.ParseOption{tdconv.StartRow(2), tdconv.TableNamePos(3, "C")},
			errMsg:   "Table name row must be smaller than the start row",
		},
		{
			caseName: "failure: KeyNameFunc",
			opts:     []tdconv.ParseOption{tdconv.KeyNameFunc(nil)},
			errMsg:   "Key name function must not be nil",
		},
	}

	for _, c := range cases {
		t.Run(c.caseName, func(t *testing.T) {
			_, err := tdconv.NewParser(c.opts...)

			if c.errMsg == "" {
				if err != nil {
					t.Errorf("error must not occur: %v", err)
					return
				}
			} else {
				if err == nil {
					t.Errorf("error must occur")
					return
				}
				if endIndex := strings.Index(err.Error(), ":"); endIndex < 0 {
					if err.Error() != c.errMsg {
						t.Errorf("error message doesn't match (expected=%s, actual=%s)", c.errMsg, err.Error())
						return
					}
				} else if err.Error()[:endIndex] != c.errMsg {
					t.Errorf("error message doesn't match (expected=%s, actual=%s)", c.errMsg, err.Error()[:endIndex])
					return
				}
			}
		})
	}
}

func TestParser_SetCommonColumns(t *testing.T) {

	cases := []struct {
		caseName   string
		p          *tdconv.Parser
		commons    [][]interface{}
		doubleCall bool
		errMsg     string
	}{
		{
			caseName: "success:nil parser",
			p:        nil,
		},
		{
			caseName: "success:default parser",
			p:        mustNewParser(),
			commons: [][]interface{}{
				row(t, "1", "created_at", "TIMESTAMP NULL", "no", "no", "no", "no", "DEFAULT CURRENT_TIMESTAMP", ""),
				row(t, "2", "updated_at", "TIMESTAMP NULL", "no", "no", "no", "no", "DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP", ""),
				row(t, "3", "deleted_at", "TIMESTAMP NULL", "no", "no", "no", "no", "", ""),
			},
		},
		{
			caseName: "success:change start row",
			p:        mustNewParser(tdconv.StartRow(5)),
			commons: [][]interface{}{
				row(t, "1", "created_at", "TIMESTAMP NULL", "no", "no", "no", "no", "DEFAULT CURRENT_TIMESTAMP", ""),
				row(t, "2", "updated_at", "TIMESTAMP NULL", "no", "no", "no", "no", "DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP", ""),
				row(t, "3", "deleted_at", "TIMESTAMP NULL", "no", "no", "no", "no", "", ""),
			},
		},
		{
			caseName: "failure:double call",
			p:        mustNewParser(),
			commons: [][]interface{}{
				row(t, "1", "created_at", "TIMESTAMP NULL", "no", "no", "no", "no", "DEFAULT CURRENT_TIMESTAMP", ""),
				row(t, "2", "updated_at", "TIMESTAMP NULL", "no", "no", "no", "no", "DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP", ""),
				row(t, "3", "deleted_at", "TIMESTAMP NULL", "no", "no", "no", "no", "", ""),
			},
			doubleCall: true,
			errMsg:     "The common columns are already set",
		},
		{
			caseName: "failure:has PK",
			p:        mustNewParser(),
			commons: [][]interface{}{
				row(t, "1", "created_at", "TIMESTAMP NULL", "yes", "no", "no", "no", "DEFAULT CURRENT_TIMESTAMP", ""),
				row(t, "2", "updated_at", "TIMESTAMP NULL", "no", "no", "no", "no", "DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP", ""),
				row(t, "3", "deleted_at", "TIMESTAMP NULL", "no", "no", "no", "no", "", ""),
			},
			errMsg: "The common column must not be PK",
		},
		{
			caseName: "failure:has index",
			p:        mustNewParser(),
			commons: [][]interface{}{
				row(t, "1", "created_at", "TIMESTAMP NULL", "no", "no", "no", "yes", "DEFAULT CURRENT_TIMESTAMP", ""),
				row(t, "2", "updated_at", "TIMESTAMP NULL", "no", "no", "no", "no", "DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP", ""),
				row(t, "3", "deleted_at", "TIMESTAMP NULL", "no", "no", "no", "no", "", ""),
			},
			errMsg: "The common column must not have index",
		},
	}

	for _, c := range cases {
		t.Run(c.caseName, func(t *testing.T) {
			cs := sheet(t, c.p, "", c.commons...)
			err := c.p.SetCommonColumns(cs)
			if c.doubleCall {
				err = c.p.SetCommonColumns(cs)
			}

			if c.errMsg == "" {
				if err != nil {
					t.Errorf("error must not occur: %v", err)
					return
				}
			} else {
				if err == nil {
					t.Errorf("error must occur")
					return
				}
				if endIndex := strings.Index(err.Error(), ":"); endIndex < 0 {
					if err.Error() != c.errMsg {
						t.Errorf("error message doesn't match (expected=%s, actual=%s)", c.errMsg, err.Error())
						return
					}
				} else if err.Error()[:endIndex] != c.errMsg {
					t.Errorf("error message doesn't match (expected=%s, actual=%s)", c.errMsg, err.Error()[:endIndex])
					return
				}
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {

	cases := []struct {
		caseName      string
		p             *tdconv.Parser
		tableName     string
		rows, commons [][]interface{}
		expected      *tdconv.Table
		errMsg        string
	}{
		{
			caseName: "success:nil parser",
			p:        nil,
			expected: nil,
		},
		{
			caseName:  "success:default parser",
			p:         mustNewParser(),
			tableName: "sample_table",
			rows: [][]interface{}{
				row(t, "1", "id", "INT UNSIGNED", "yes", "yes", "no", "no", "AUTO_INCREMENT", "this is id!"),
				row(t, "2", "foo", "VARCHAR(32)", "no", "yes", "yes", "no", "", ""),
				row(t, "3", "abc", "", "no", "no", "no", "no", "", "this row is ignored"),
				row(t, "4", "bar", "VARCHAR(32)", "no", "no", "no", "yes", "", ""),
				row(t, "", "xyz", "VARCHAR(32)", "no", "no", "no", "no", "", "this row is the end"),
				row(t, "6", "zzz", "VARCHAR(32)", "no", "no", "no", "no", "", "this row is unreachable"),
			},
			expected: &tdconv.Table{
				Name: "sample_table",
				Columns: []tdconv.Column{
					{Name: "id", Type: "INT UNSIGNED", PKey: true, NotNull: true, Unique: false, Index: false, Option: "AUTO_INCREMENT", Comment: "this is id!", IsCommon: false},
					{Name: "foo", Type: "VARCHAR(32)", PKey: false, NotNull: true, Unique: true, Index: false, Option: "", Comment: "", IsCommon: false},
					{Name: "bar", Type: "VARCHAR(32)", PKey: false, NotNull: false, Unique: false, Index: true, Option: "", Comment: "", IsCommon: false},
				},
				PKeyColumns: []string{"id"},
				UniqueKeys:  nil, // single unique key is not stored in this slice
				IndexKeys:   []tdconv.Key{{Name: "bar_key", Columns: []string{"bar"}}},
			},
		},
		{
			caseName:  "success:change table name position",
			p:         mustNewParser(tdconv.TableNamePos(0, "A")),
			tableName: "sample_table",
			rows: [][]interface{}{
				row(t, "1", "id", "INT UNSIGNED", "yes", "yes", "no", "no", "AUTO_INCREMENT", "this is id!"),
				row(t, "2", "foo", "VARCHAR(32)", "no", "yes", "yes", "no", "", ""),
				row(t, "3", "bar", "VARCHAR(32)", "no", "no", "no", "yes", "", ""),
			},
			expected: &tdconv.Table{
				Name: "sample_table",
				Columns: []tdconv.Column{
					{Name: "id", Type: "INT UNSIGNED", PKey: true, NotNull: true, Unique: false, Index: false, Option: "AUTO_INCREMENT", Comment: "this is id!", IsCommon: false},
					{Name: "foo", Type: "VARCHAR(32)", PKey: false, NotNull: true, Unique: true, Index: false, Option: "", Comment: "", IsCommon: false},
					{Name: "bar", Type: "VARCHAR(32)", PKey: false, NotNull: false, Unique: false, Index: true, Option: "", Comment: "", IsCommon: false},
				},
				PKeyColumns: []string{"id"},
				UniqueKeys:  nil, // single unique key is not stored in this slice
				IndexKeys:   []tdconv.Key{{Name: "bar_key", Columns: []string{"bar"}}},
			},
		},
		{
			caseName:  "success:change start row",
			p:         mustNewParser(tdconv.StartRow(5)),
			tableName: "sample_table",
			rows: [][]interface{}{
				row(t, "1", "id", "INT UNSIGNED", "yes", "yes", "no", "no", "AUTO_INCREMENT", "this is id!"),
				row(t, "2", "foo", "VARCHAR(32)", "no", "yes", "yes", "no", "", ""),
				row(t, "3", "bar", "VARCHAR(32)", "no", "no", "no", "yes", "", ""),
			},
			expected: &tdconv.Table{
				Name: "sample_table",
				Columns: []tdconv.Column{
					{Name: "id", Type: "INT UNSIGNED", PKey: true, NotNull: true, Unique: false, Index: false, Option: "AUTO_INCREMENT", Comment: "this is id!", IsCommon: false},
					{Name: "foo", Type: "VARCHAR(32)", PKey: false, NotNull: true, Unique: true, Index: false, Option: "", Comment: "", IsCommon: false},
					{Name: "bar", Type: "VARCHAR(32)", PKey: false, NotNull: false, Unique: false, Index: true, Option: "", Comment: "", IsCommon: false},
				},
				PKeyColumns: []string{"id"},
				UniqueKeys:  nil, // single unique key is not stored in this slice
				IndexKeys:   []tdconv.Key{{Name: "bar_key", Columns: []string{"bar"}}},
			},
		},
		{
			caseName:  "success:change bool string",
			p:         mustNewParser(tdconv.BoolString("OK")),
			tableName: "sample_table",
			rows: [][]interface{}{
				row(t, "1", "id", "INT UNSIGNED", "OK", "OK", "yes", "yes", "AUTO_INCREMENT", "this is id!"),
				row(t, "2", "foo", "VARCHAR(32)", "yes", "OK", "OK", "yes", "", ""),
				row(t, "3", "bar", "VARCHAR(32)", "yes", "yes", "yes", "OK", "", ""),
			},
			expected: &tdconv.Table{
				Name: "sample_table",
				Columns: []tdconv.Column{
					{Name: "id", Type: "INT UNSIGNED", PKey: true, NotNull: true, Unique: false, Index: false, Option: "AUTO_INCREMENT", Comment: "this is id!", IsCommon: false},
					{Name: "foo", Type: "VARCHAR(32)", PKey: false, NotNull: true, Unique: true, Index: false, Option: "", Comment: "", IsCommon: false},
					{Name: "bar", Type: "VARCHAR(32)", PKey: false, NotNull: false, Unique: false, Index: true, Option: "", Comment: "", IsCommon: false},
				},
				PKeyColumns: []string{"id"},
				UniqueKeys:  nil, // single unique key is not stored in this slice
				IndexKeys:   []tdconv.Key{{Name: "bar_key", Columns: []string{"bar"}}},
			},
		},
		{
			caseName:  "success:change key name funtion",
			p:         mustNewParser(tdconv.KeyNameFunc(func(s string) string { return "key_" + s })),
			tableName: "sample_table",
			rows: [][]interface{}{
				row(t, "1", "id", "INT UNSIGNED", "yes", "yes", "no", "no", "AUTO_INCREMENT", "this is id!"),
				row(t, "2", "foo", "VARCHAR(32)", "no", "yes", "yes", "no", "", ""),
				row(t, "3", "bar", "VARCHAR(32)", "no", "no", "no", "yes", "", ""),
			},
			expected: &tdconv.Table{
				Name: "sample_table",
				Columns: []tdconv.Column{
					{Name: "id", Type: "INT UNSIGNED", PKey: true, NotNull: true, Unique: false, Index: false, Option: "AUTO_INCREMENT", Comment: "this is id!", IsCommon: false},
					{Name: "foo", Type: "VARCHAR(32)", PKey: false, NotNull: true, Unique: true, Index: false, Option: "", Comment: "", IsCommon: false},
					{Name: "bar", Type: "VARCHAR(32)", PKey: false, NotNull: false, Unique: false, Index: true, Option: "", Comment: "", IsCommon: false},
				},
				PKeyColumns: []string{"id"},
				UniqueKeys:  nil, // single unique key is not stored in this slice
				IndexKeys:   []tdconv.Key{{Name: "key_bar", Columns: []string{"bar"}}},
			},
		},
		{
			caseName:  "success:common columns",
			p:         mustNewParser(),
			tableName: "sample_table",
			rows: [][]interface{}{
				row(t, "1", "id", "INT UNSIGNED", "yes", "yes", "no", "no", "AUTO_INCREMENT", "this is id!"),
				row(t, "2", "foo", "VARCHAR(32)", "no", "yes", "yes", "no", "", ""),
				row(t, "3", "bar", "VARCHAR(32)", "no", "no", "no", "yes", "", ""),
			},
			commons: [][]interface{}{
				row(t, "1", "created_at", "TIMESTAMP NULL", "no", "no", "no", "no", "DEFAULT CURRENT_TIMESTAMP", ""),
				row(t, "2", "updated_at", "TIMESTAMP NULL", "no", "no", "no", "no", "DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP", ""),
				row(t, "3", "deleted_at", "TIMESTAMP NULL", "no", "no", "no", "no", "", ""),
			},
			expected: &tdconv.Table{
				Name: "sample_table",
				Columns: []tdconv.Column{
					{Name: "id", Type: "INT UNSIGNED", PKey: true, NotNull: true, Unique: false, Index: false, Option: "AUTO_INCREMENT", Comment: "this is id!", IsCommon: false},
					{Name: "foo", Type: "VARCHAR(32)", PKey: false, NotNull: true, Unique: true, Index: false, Option: "", Comment: "", IsCommon: false},
					{Name: "bar", Type: "VARCHAR(32)", PKey: false, NotNull: false, Unique: false, Index: true, Option: "", Comment: "", IsCommon: false},
					{Name: "created_at", Type: "TIMESTAMP NULL", PKey: false, NotNull: false, Unique: false, Index: false, Option: "DEFAULT CURRENT_TIMESTAMP", Comment: "", IsCommon: true},
					{Name: "updated_at", Type: "TIMESTAMP NULL", PKey: false, NotNull: false, Unique: false, Index: false, Option: "DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP", Comment: "", IsCommon: true},
					{Name: "deleted_at", Type: "TIMESTAMP NULL", PKey: false, NotNull: false, Unique: false, Index: false, Option: "", Comment: "", IsCommon: true},
				},
				PKeyColumns: []string{"id"},
				UniqueKeys:  nil, // single unique key is not stored in this slice
				IndexKeys:   []tdconv.Key{{Name: "bar_key", Columns: []string{"bar"}}},
			},
		},
		{
			caseName:  "failure:no table name",
			p:         mustNewParser(),
			tableName: "",
			rows: [][]interface{}{
				row(t, "1", "id", "INT UNSIGNED", "yes", "yes", "no", "no", "AUTO_INCREMENT", "this is id!"),
				row(t, "2", "foo", "VARCHAR(32)", "no", "yes", "yes", "no", "", ""),
				row(t, "3", "bar", "VARCHAR(32)", "no", "no", "no", "yes", "", ""),
			},
			errMsg: "Table name is required",
		},
		{
			caseName:  "failure:no columns",
			p:         mustNewParser(),
			tableName: "sample_table",
			rows:      [][]interface{}{},
			errMsg:    "The length of table columns must not be zero",
		},
	}

	for _, c := range cases {
		t.Run(c.caseName, func(t *testing.T) {
			if c.commons != nil {
				cs := sheet(t, c.p, "", c.commons...)
				if err := c.p.SetCommonColumns(cs); err != nil {
					t.Errorf("error must not occur at SetCommonColumns: %v", err)
					return
				}
			}
			s := sheet(t, c.p, c.tableName, c.rows...)
			tb, err := c.p.Parse(s)

			if c.errMsg == "" {
				if err != nil {
					t.Errorf("error must not occur: %v", err)
					return
				}
				if !reflect.DeepEqual(tb, c.expected) {
					t.Errorf("value doesn't match (expected=%s, actual=%s)", gostr.Stringify(c.expected), gostr.Stringify(tb))
					return
				}
			} else {
				if err == nil {
					t.Errorf("error must occur")
					return
				}
				if endIndex := strings.Index(err.Error(), ":"); endIndex < 0 {
					if err.Error() != c.errMsg {
						t.Errorf("error message doesn't match (expected=%s, actual=%s)", c.errMsg, err.Error())
						return
					}
				} else if err.Error()[:endIndex] != c.errMsg {
					t.Errorf("error message doesn't match (expected=%s, actual=%s)", c.errMsg, err.Error()[:endIndex])
					return
				}
			}
		})
	}
}

func sheet(t *testing.T, p *tdconv.Parser, tableName string, rows ...[]interface{}) *gsheets.Sheet {
	t.Helper()
	var header [][]interface{}
	for r := 0; r < p.StartRow(); r++ {
		if r == p.TableNameRow() {
			var tableNameRow []interface{}
			for c := 0; c < p.TableNameColumn(); c++ {
				tableNameRow = append(tableNameRow, "")
			}
			header = append(header, append(tableNameRow, tableName))
		} else {
			header = append(header, []interface{}{})
		}
	}
	return gsheets.NewSheet(t, append(header, rows...))
}

func row(t *testing.T, no, name, typ, pk, notNull, unique, index, option, comment string) []interface{} {
	t.Helper()
	return []interface{}{"", no, name, typ, pk, notNull, unique, index, option, comment}
}
