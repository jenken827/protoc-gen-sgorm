package gen

import "testing"

func TestTagString(t *testing.T) {
	cases := []struct {
		name string
		tag  tagSpec
		want string
	}{
		{name: "空", tag: tagSpec{}, want: ""},
		{
			name: "主键自增",
			tag:  tagSpec{primaryKey: true, autoIncr: true},
			want: "primaryKey;autoIncrement",
		},
		{
			name: "全字段按固定顺序",
			tag: tagSpec{
				primaryKey: true, autoIncr: true, notNull: true, unique: true,
				column: "provider_type", typ: "uuid", size: 36, precision: 2,
				def: "''", index: "idx_a", uniqueIndex: "idx_b",
				autoCreate: true, autoUpdate: true, serializer: "json",
			},
			want: "primaryKey;autoIncrement;not null;unique;column:provider_type;type:uuid;" +
				"size:36;precision:2;default:'';index:idx_a;uniqueIndex:idx_b;" +
				"autoCreateTime;autoUpdateTime;serializer:json",
		},
		{
			name: "索引带参数",
			tag:  tagSpec{index: "idx_f_npoi,length:10"},
			want: "index:idx_f_npoi,length:10",
		},
	}
	for _, c := range cases {
		if got := c.tag.String(); got != c.want {
			t.Errorf("%s: tag = %q, 期望 %q", c.name, got, c.want)
		}
	}
}
