package gen

import "google.golang.org/protobuf/compiler/protogen"

// emitFromField 生成 From<Msg> 里的单字段赋值（proto → 实体）。
func emitFromField(g *protogen.GeneratedFile, f field) {
	dst := "out." + f.goName
	src := "in." + f.goName
	switch f.conv {
	case convAssign:
		g.P(dst, " = ", src)
	case convEnum:
		if f.pointer {
			g.P(dst, " = (*int32)(", src, ")")
		} else {
			g.P(dst, " = int32(", src, ")")
		}
	case convTimestamp:
		g.P("if ", src, " != nil {")
		g.P("v := ", src, ".AsTime()")
		g.P(dst, " = &v")
		g.P("}")
	case convDeletedAt:
		g.P("if ", src, " != nil {")
		g.P(dst, " = &", g.QualifiedGoIdent(gormDeleted), "{Time: ", src, ".AsTime(), Valid: true}")
		g.P("}")
	}
}

// emitToPBField 生成 ToPB 里的单字段赋值（实体 → proto）。
func emitToPBField(g *protogen.GeneratedFile, f field) {
	dst := "out." + f.goName
	src := "m." + f.goName
	switch f.conv {
	case convAssign:
		g.P(dst, " = ", src)
	case convEnum:
		if f.pointer {
			g.P(dst, " = (*", g.QualifiedGoIdent(*f.enumPtr), ")(", src, ")")
		} else {
			g.P(dst, " = ", g.QualifiedGoIdent(*f.enumPtr), "(", src, ")")
		}
	case convTimestamp:
		g.P("if ", src, " != nil {")
		g.P(dst, " = ", g.QualifiedGoIdent(pbTimestamp), "(*", src, ")")
		g.P("}")
	case convDeletedAt:
		g.P("if ", src, " != nil && ", src, ".Valid {")
		g.P(dst, " = ", g.QualifiedGoIdent(pbTimestamp), "(", src, ".Time)")
		g.P("}")
	}
}
