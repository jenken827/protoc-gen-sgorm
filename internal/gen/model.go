package gen

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/jenken827/protoc-gen-sgorm/gen/sgorm"
)

// 生成代码里用到的外部类型（惰性注册 import）。
var (
	contextIdent = protogen.GoIdent{GoName: "Context", GoImportPath: "context"}
	timeType     = protogen.GoIdent{GoName: "Time", GoImportPath: "time"}
	gormDeleted  = protogen.GoIdent{GoName: "DeletedAt", GoImportPath: "gorm.io/gorm"}
	pbTimestamp  = protogen.GoIdent{GoName: "New", GoImportPath: "google.golang.org/protobuf/types/known/timestamppb"}
)

// convKind 决定 ToPB / From 里的赋值写法。
type convKind int

const (
	convAssign    convKind = iota // 两边类型一致，直接赋值
	convEnum                      // 枚举指针 ↔ *int32
	convTimestamp                 // *timestamppb.Timestamp ↔ *time.Time
	convDeletedAt                 // *timestamppb.Timestamp ↔ *gorm.DeletedAt
)

// typeRef 描述一个 Go 类型，渲染时才注册 import。
type typeRef struct {
	slice   bool
	pointer bool
	builtin string // 内置类型名；与 ident 二选一
	ident   *protogen.GoIdent
}

func (t typeRef) render(g *protogen.GeneratedFile) string {
	out := ""
	if t.slice {
		out += "[]"
	}
	if t.pointer {
		out += "*"
	}
	if t.ident != nil {
		out += g.QualifiedGoIdent(*t.ident)
	} else {
		out += t.builtin
	}
	return out
}

// field 是一个 ORM 字段。
type field struct {
	goName  string            // Go 字段名，与 proto 生成的一致（Id / UserId）
	pbName  string            // proto 字段名，仅用于文档
	ormType typeRef           // 实体里的类型
	conv    convKind          // 转换方式
	enumPtr *protogen.GoIdent // 枚举类型（convEnum 时使用）
	pointer bool              // 原 proto 字段是否有 presence（决定枚举要不要指针）
	tag     string            // 渲染好的 gorm tag
	comment string            // 已带 // 前缀的多行注释
}

// model 是一个待生成的消息。
type model struct {
	pbName  string
	ormName string
	pbIdent protogen.GoIdent
	table   string
	doc     string
	fields  []field
}

// buildModel 把 ormable 消息转成 model；非 ormable 返回 nil。
func buildModel(msg *protogen.Message) (*model, error) {
	opts := messageOpts(msg)
	if !opts.GetOrmable() {
		return nil, nil
	}
	m := &model{
		pbName:  msg.GoIdent.GoName,
		ormName: msg.GoIdent.GoName + "ORM",
		pbIdent: msg.GoIdent,
		table:   tableName(msg.GoIdent.GoName, opts.GetTable()),
		doc:     commentText(msg.Comments),
	}
	for _, f := range msg.Fields {
		fld, ok := buildField(f)
		if !ok {
			continue
		}
		m.fields = append(m.fields, fld)
	}
	return m, nil
}

// buildField 把 proto 字段转成 ORM 字段；跳过 repeated / map / 未支持的嵌套消息。
func buildField(f *protogen.Field) (field, bool) {
	opts := fieldOpts(f)
	if opts.GetIgnore() {
		return field{}, false
	}
	// repeated / map 在实体里没有对应列（关联表、JSON 列都不是自动能推的）。
	if f.Desc.IsList() || f.Desc.IsMap() {
		return field{}, false
	}

	tag := tagSpecFrom(opts.GetTag())
	out := field{
		goName:  f.GoName,
		pbName:  string(f.Desc.Name()),
		pointer: f.Desc.HasPresence(),
		comment: commentText(f.Comments),
	}

	switch {
	case isTimestamp(f):
		if tag.typ == "deleted_at" {
			tag.typ = "timestamp"
			out.conv = convDeletedAt
			out.ormType = typeRef{pointer: true, ident: &gormDeleted}
		} else {
			out.conv = convTimestamp
			out.ormType = typeRef{pointer: true, ident: &timeType}
		}
	case f.Enum != nil:
		ident := f.Enum.GoIdent
		out.conv = convEnum
		out.enumPtr = &ident
		out.ormType = typeRef{pointer: out.pointer, builtin: "int32"}
	case f.Message != nil:
		// 其它消息类型（嵌套 DTO / 关联对象）不落列，需要时用 serializer 或另建表。
		return field{}, false
	case f.Desc.Kind() == protoreflect.BytesKind:
		out.conv = convAssign
		out.ormType = typeRef{slice: true, builtin: "byte"}
	default:
		base, err := scalarGoType(f.Desc.Kind())
		if err != nil {
			return field{}, false
		}
		out.conv = convAssign
		out.ormType = typeRef{pointer: out.pointer, builtin: base}
	}

	out.tag = tag.String()
	return out, true
}

// scalarGoType 返回标量对应的 Go 内置类型名。
func scalarGoType(kind protoreflect.Kind) (string, error) {
	switch kind {
	case protoreflect.StringKind:
		return "string", nil
	case protoreflect.BoolKind:
		return "bool", nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32", nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32", nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64", nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64", nil
	case protoreflect.FloatKind:
		return "float32", nil
	case protoreflect.DoubleKind:
		return "float64", nil
	}
	return "", fmt.Errorf("不支持的标量类型: %v", kind)
}

func isTimestamp(f *protogen.Field) bool {
	return f.Message != nil && f.Message.Desc.FullName() == "google.protobuf.Timestamp"
}

// messageOpts 读消息上的 (sgorm.opts)。
func messageOpts(msg *protogen.Message) *sgorm.Opts {
	opts, ok := msg.Desc.Options().(*descriptorpb.MessageOptions)
	if !ok || opts == nil || !proto.HasExtension(opts, sgorm.E_Opts) {
		return nil
	}
	v, ok := proto.GetExtension(opts, sgorm.E_Opts).(*sgorm.Opts)
	if !ok {
		return nil
	}
	return v
}

// fieldOpts 读字段上的 (sgorm.field)。
func fieldOpts(f *protogen.Field) *sgorm.Field {
	opts, ok := f.Desc.Options().(*descriptorpb.FieldOptions)
	if !ok || opts == nil || !proto.HasExtension(opts, sgorm.E_Field) {
		return nil
	}
	v, ok := proto.GetExtension(opts, sgorm.E_Field).(*sgorm.Field)
	if !ok {
		return nil
	}
	return v
}

// pbType 返回 proto 消息在生成文件里的限定名（native.Transaction）。
func pbType(g *protogen.GeneratedFile, m *model) string {
	return g.QualifiedGoIdent(m.pbIdent)
}

// ctxType 渲染 context.Context。
func ctxType(g *protogen.GeneratedFile) string {
	return g.QualifiedGoIdent(contextIdent)
}

// commentText 取字段/消息注释：优先前导（写在声明上一行），退回行尾（写在声明右侧）。
// proto 里两种写法都常见，注释是实体字段文档的唯一来源，两种都要透传。
func commentText(c protogen.CommentSet) string {
	leading, trailing := c.Leading.String(), c.Trailing.String()
	switch {
	case strings.TrimSpace(leading) == "":
		return trailing
	case strings.TrimSpace(trailing) == "":
		return leading
	default:
		return leading + "//\n" + trailing
	}
}
