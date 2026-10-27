package gen

import "testing"

func TestSnake(t *testing.T) {
	cases := map[string]string{
		"Transaction":     "transaction",
		"AssetAccount":    "asset_account",
		"WebdavAccount":   "webdav_account",
		"DealEntry":       "deal_entry",
		"UserSpace":       "user_space",
		"FileHistory":     "file_history",
		"HTMLParserTask":  "html_parser_task",
		"UploadChunk2Rec": "upload_chunk2_rec",
	}
	for in, want := range cases {
		if got := snake(in); got != want {
			t.Errorf("snake(%q) = %q, 期望 %q", in, got, want)
		}
	}
}

func TestPluralize(t *testing.T) {
	cases := map[string]string{
		"transaction":      "transactions",
		"category":         "categories",
		"history":          "histories",
		"file":             "files",
		"tag":              "tags",
		"box":              "boxes",
		"branch":           "branches",
		"bus":              "buses",
		"day":              "days",
		"user_space":       "user_spaces",
		"storage_provider": "storage_providers",
	}
	for in, want := range cases {
		if got := pluralize(in); got != want {
			t.Errorf("pluralize(%q) = %q, 期望 %q", in, got, want)
		}
	}
}

func TestTableName(t *testing.T) {
	if got := tableName("Category", ""); got != "categories" {
		t.Errorf("默认表名 = %q, 期望 categories", got)
	}
	if got := tableName("Person", "people"); got != "people" {
		t.Errorf("覆盖表名 = %q, 期望 people", got)
	}
}
