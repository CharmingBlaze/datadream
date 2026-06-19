package codegen

import (
	"fmt"

	"datadream/internal/ast"
)

func (g *Generator) emitSaveStructStubs(s *ast.StructDecl) {
	g.emit("\nstatic void %s_serialize(const %s* src, FILE* out) {\n", s.Name, s.Name)
	g.indent++
	g.iemit("if (!src || !out) return;\n")
	for _, f := range s.Fields {
		g.emitSaveFieldWrite(s, f, "src", "out")
	}
	g.indent--
	g.emit("}\n")

	g.emit("\nstatic void %s_deserialize(%s* dst, FILE* in) {\n", s.Name, s.Name)
	g.indent++
	g.iemit("if (!dst || !in) return;\n")
	for _, f := range s.Fields {
		g.emitSaveFieldRead(s, f, "dst", "in")
	}
	g.indent--
	g.emit("}\n")
}

func (g *Generator) emitSaveFieldWrite(s *ast.StructDecl, f ast.FieldDecl, obj, out string) {
	cType := "int"
	if f.Type != nil {
		cType = g.typeToC(f.Type)
	}
	switch cType {
	case "int", "float", "bool", "char", "unsigned char", "double", "short", "long long":
		g.iemit("fwrite(&%s->%s, sizeof(%s), 1, %s);\n", obj, f.Name, cType, out)
	case "Vec2":
		g.iemit("fwrite(&%s->%s.x, sizeof(float), 1, %s);\n", obj, f.Name, out)
		g.iemit("fwrite(&%s->%s.y, sizeof(float), 1, %s);\n", obj, f.Name, out)
	case "Vec3":
		g.iemit("fwrite(&%s->%s.x, sizeof(float), 1, %s);\n", obj, f.Name, out)
		g.iemit("fwrite(&%s->%s.y, sizeof(float), 1, %s);\n", obj, f.Name, out)
		g.iemit("fwrite(&%s->%s.z, sizeof(float), 1, %s);\n", obj, f.Name, out)
	case "Vec4":
		g.iemit("fwrite(&%s->%s.x, sizeof(float), 1, %s);\n", obj, f.Name, out)
		g.iemit("fwrite(&%s->%s.y, sizeof(float), 1, %s);\n", obj, f.Name, out)
		g.iemit("fwrite(&%s->%s.z, sizeof(float), 1, %s);\n", obj, f.Name, out)
		g.iemit("fwrite(&%s->%s.w, sizeof(float), 1, %s);\n", obj, f.Name, out)
	case "char*":
		g.iemit("{\n")
		g.indent++
		g.iemit("const char* _s = %s->%s ? %s->%s : \"\";\n", obj, f.Name, obj, f.Name)
		g.iemit("int _len = (int)strlen(_s);\n")
		g.iemit("fwrite(&_len, sizeof(int), 1, %s);\n", out)
		g.iemit("if (_len > 0) fwrite(_s, 1, (size_t)_len, %s);\n", out)
		g.indent--
		g.iemit("}\n")
	default:
		g.addErrorAt(s.Pos(), fmt.Sprintf("@save: unsupported field type %q for %s.%s", cType, s.Name, f.Name),
			"@save supports int, float, bool, Vec2/3/4, and string fields")
	}
}

func (g *Generator) emitSaveFieldRead(s *ast.StructDecl, f ast.FieldDecl, obj, in string) {
	cType := "int"
	if f.Type != nil {
		cType = g.typeToC(f.Type)
	}
	switch cType {
	case "int", "float", "bool", "char", "unsigned char", "double", "short", "long long":
		g.iemit("fread(&%s->%s, sizeof(%s), 1, %s);\n", obj, f.Name, cType, in)
	case "Vec2":
		g.iemit("fread(&%s->%s.x, sizeof(float), 1, %s);\n", obj, f.Name, in)
		g.iemit("fread(&%s->%s.y, sizeof(float), 1, %s);\n", obj, f.Name, in)
	case "Vec3":
		g.iemit("fread(&%s->%s.x, sizeof(float), 1, %s);\n", obj, f.Name, in)
		g.iemit("fread(&%s->%s.y, sizeof(float), 1, %s);\n", obj, f.Name, in)
		g.iemit("fread(&%s->%s.z, sizeof(float), 1, %s);\n", obj, f.Name, in)
	case "Vec4":
		g.iemit("fread(&%s->%s.x, sizeof(float), 1, %s);\n", obj, f.Name, in)
		g.iemit("fread(&%s->%s.y, sizeof(float), 1, %s);\n", obj, f.Name, in)
		g.iemit("fread(&%s->%s.z, sizeof(float), 1, %s);\n", obj, f.Name, in)
		g.iemit("fread(&%s->%s.w, sizeof(float), 1, %s);\n", obj, f.Name, in)
	case "char*":
		g.iemit("{\n")
		g.indent++
		g.iemit("int _len = 0;\n")
		g.iemit("fread(&_len, sizeof(int), 1, %s);\n", in)
		g.iemit("if (_len < 0) _len = 0;\n")
		g.iemit("free(%s->%s);\n", obj, f.Name)
		g.iemit("%s->%s = (char*)calloc((size_t)_len + 1, 1);\n", obj, f.Name)
		g.iemit("if (_len > 0) fread(%s->%s, 1, (size_t)_len, %s);\n", obj, f.Name, in)
		g.iemit("%s->%s[_len] = '\\0';\n", obj, f.Name)
		g.indent--
		g.iemit("}\n")
	default:
		g.addErrorAt(s.Pos(), fmt.Sprintf("@save: unsupported field type %q for %s.%s", cType, s.Name, f.Name),
			"@save supports int, float, bool, Vec2/3/4, and string fields")
	}
}

func (g *Generator) isPackedEntity(name string) bool {
	return g.packedEntities != nil && g.packedEntities[name]
}

func (g *Generator) packedPoolType(entity string) string {
	return entity + "Pool"
}

func (g *Generator) packedPoolVar(entity string) string {
	return entity + "_pool"
}

func (g *Generator) emitPackedEntityFieldAccess(entity, field string) bool {
	if !g.isPackedEntity(entity) {
		return false
	}
	pool := g.packedPoolVar(entity)
	switch field {
	case "position":
		g.emit("%s.positions[self->idx]", pool)
	case "velocity":
		g.emit("%s.velocities[self->idx]", pool)
	case "active":
		g.emit("%s.actives[self->idx]", pool)
	default:
		g.emit("%s.%s[self->idx]", pool, field)
	}
	return true
}

func (g *Generator) emitPackedEntityVarFieldAccess(varName, entity, field string) bool {
	if !g.isPackedEntity(entity) {
		return false
	}
	pool := g.packedPoolVar(entity)
	switch field {
	case "position":
		g.emit("%s.positions[%s->idx]", pool, varName)
	case "velocity":
		g.emit("%s.velocities[%s->idx]", pool, varName)
	case "active":
		g.emit("%s.actives[%s->idx]", pool, varName)
	default:
		g.emit("%s.%s[%s->idx]", pool, field, varName)
	}
	return true
}

func (g *Generator) emitPackedEntityFieldAssign(entity, field string) {
	pool := g.packedPoolVar(entity)
	switch field {
	case "position":
		g.emit("%s.positions[self->idx]", pool)
	case "velocity":
		g.emit("%s.velocities[self->idx]", pool)
	case "active":
		g.emit("%s.actives[self->idx]", pool)
	default:
		g.emit("%s.%s[self->idx]", pool, field)
	}
}
