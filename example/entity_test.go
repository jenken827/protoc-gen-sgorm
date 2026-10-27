package example_test

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm/schema"

	"github.com/jenken827/protoc-gen-sgorm/example/gen/entity"
	"github.com/jenken827/protoc-gen-sgorm/example/gen/native"
)

// TestSchema 用 GORM 自己的解析器验证生成的 tag 能被正确理解（不是只看文本）。
func TestSchema(t *testing.T) {
	cache := &sync.Map{}

	s, err := schema.Parse(&entity.TransactionORM{}, cache, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("解析 TransactionORM 失败: %v", err)
	}
	if s.Table != "transactions" {
		t.Errorf("表名 = %q, 期望 transactions", s.Table)
	}
	if s.PrioritizedPrimaryField == nil || s.PrioritizedPrimaryField.DBName != "id" {
		t.Errorf("主键未按 tag 识别: %+v", s.PrioritizedPrimaryField)
	}
	if f := s.LookUpField("UserId"); f == nil || f.DBName != "owner_id" {
		t.Errorf("column 改写未生效: %+v", f)
	}
	if f := s.LookUpField("DeletedAt"); f == nil || f.IndirectFieldType.String() != "gorm.DeletedAt" {
		t.Errorf("软删字段类型不对: %+v", f)
	}
	if s.LookUpField("Note") != nil {
		t.Error("ignore 的字段仍生成了列")
	}
	if s.LookUpField("TagIds") != nil {
		t.Error("repeated 字段不应生成列")
	}

	f, err := schema.Parse(&entity.FileORM{}, cache, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("解析 FileORM 失败: %v", err)
	}
	if f.Table != "files" {
		t.Errorf("File 表名 = %q, 期望 files", f.Table)
	}
	if id := f.LookUpField("Id"); id == nil || id.TagSettings["TYPE"] != "uuid" {
		t.Errorf("type:uuid 未生效: %+v", id)
	}

	p, err := schema.Parse(&entity.PersonORM{}, cache, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("解析 PersonORM 失败: %v", err)
	}
	if p.Table != "people" {
		t.Errorf("table 覆盖未生效: %q", p.Table)
	}
}

// TestRoundTrip 验证 From<Msg> 与 ToPB 往返一致。
func TestRoundTrip(t *testing.T) {
	ctx := context.Background()
	created := timestamppb.New(time.Date(2026, 10, 27, 8, 30, 0, 123456789, time.UTC))
	updated := timestamppb.New(time.Date(2026, 10, 27, 9, 0, 0, 0, time.UTC))
	deleted := timestamppb.New(time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC))
	in := &native.Transaction{
		Id:        proto.Uint32(7),
		CreatedAt: created,
		UpdatedAt: updated,
		DeletedAt: deleted,
		Amount:    proto.Int64(12345),
		Type:      native.TransactionType_INCOME.Enum(),
		UserId:    proto.String("user-1"),
	}

	orm, err := entity.FromTransaction(ctx, in)
	if err != nil {
		t.Fatalf("FromTransaction 失败: %v", err)
	}
	if orm.TableName() != "transactions" {
		t.Errorf("表名 = %q", orm.TableName())
	}
	if orm.CreatedAt == nil || !orm.CreatedAt.Equal(created.AsTime()) {
		t.Errorf("created_at 转换有误: %+v", orm.CreatedAt)
	}
	if orm.DeletedAt == nil || !orm.DeletedAt.Valid {
		t.Errorf("deleted_at 未映射成 gorm.DeletedAt: %+v", orm.DeletedAt)
	}
	if orm.Type == nil || *orm.Type != int32(native.TransactionType_INCOME) {
		t.Errorf("枚举转换有误: %+v", orm.Type)
	}

	out, err := orm.ToPB(ctx)
	if err != nil {
		t.Fatalf("ToPB 失败: %v", err)
	}
	if !proto.Equal(in, out) {
		t.Errorf("往返不一致:\n in=%v\nout=%v", in, out)
	}
}

// TestRoundTripNil 验证空指针字段不会被凭空造出来。
func TestRoundTripNil(t *testing.T) {
	ctx := context.Background()
	orm, err := entity.FromTransaction(ctx, &native.Transaction{})
	if err != nil {
		t.Fatalf("FromTransaction 失败: %v", err)
	}
	if orm.CreatedAt != nil || orm.DeletedAt != nil || orm.Type != nil {
		t.Errorf("空字段不应有值: %+v", orm)
	}
	out, err := orm.ToPB(ctx)
	if err != nil {
		t.Fatalf("ToPB 失败: %v", err)
	}
	if !proto.Equal(out, &native.Transaction{}) {
		t.Errorf("空实体应转回空消息: %v", out)
	}
}

// TestComments 验证 proto 注释透传到实体字段：行尾注释（`;  // 金额`）也要带过来。
func TestComments(t *testing.T) {
	src, err := os.ReadFile("gen/entity/transaction.sgorm.go")
	if err != nil {
		t.Fatalf("读取生成文件失败: %v", err)
	}
	if !strings.Contains(string(src), "// 金额，单位为分") {
		t.Error("行尾注释没有透传到实体字段")
	}
}
