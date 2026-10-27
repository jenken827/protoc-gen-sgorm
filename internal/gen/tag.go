package gen

import (
	"strconv"
	"strings"

	"github.com/jenken827/protoc-gen-sgorm/gen/sgorm"
)

// tagSpec 是 gorm tag 的中间表示，渲染顺序固定，便于 diff。
type tagSpec struct {
	primaryKey  bool
	autoIncr    bool
	notNull     bool
	unique      bool
	column      string
	typ         string
	size        int32
	precision   int32
	def         string
	index       string
	uniqueIndex string
	autoCreate  bool
	autoUpdate  bool
	serializer  string
}

func tagSpecFrom(t *sgorm.Tag) tagSpec {
	if t == nil {
		return tagSpec{}
	}
	return tagSpec{
		primaryKey:  t.GetPrimaryKey(),
		autoIncr:    t.GetAutoIncrement(),
		notNull:     t.GetNotNull(),
		unique:      t.GetUnique(),
		column:      t.GetColumn(),
		typ:         t.GetType(),
		size:        t.GetSize(),
		precision:   t.GetPrecision(),
		def:         t.GetDefault(),
		index:       t.GetIndex(),
		uniqueIndex: t.GetUniqueIndex(),
		autoCreate:  t.GetAutoCreateTime(),
		autoUpdate:  t.GetAutoUpdateTime(),
		serializer:  t.GetSerializer(),
	}
}

// String 按固定顺序拼出 gorm tag 值；没有设置返回空串。
func (t tagSpec) String() string {
	var parts []string
	add := func(s string) { parts = append(parts, s) }

	if t.primaryKey {
		add("primaryKey")
	}
	if t.autoIncr {
		add("autoIncrement")
	}
	if t.notNull {
		add("not null")
	}
	if t.unique {
		add("unique")
	}
	if t.column != "" {
		add("column:" + t.column)
	}
	if t.typ != "" {
		add("type:" + t.typ)
	}
	if t.size > 0 {
		add("size:" + strconv.Itoa(int(t.size)))
	}
	if t.precision > 0 {
		add("precision:" + strconv.Itoa(int(t.precision)))
	}
	if t.def != "" {
		add("default:" + t.def)
	}
	if t.index != "" {
		add("index:" + t.index)
	}
	if t.uniqueIndex != "" {
		add("uniqueIndex:" + t.uniqueIndex)
	}
	if t.autoCreate {
		add("autoCreateTime")
	}
	if t.autoUpdate {
		add("autoUpdateTime")
	}
	if t.serializer != "" {
		add("serializer:" + t.serializer)
	}
	return strings.Join(parts, ";")
}

// quote 用 Go 字符串字面量包一层。
func quote(s string) string {
	return strconv.Quote(s)
}
