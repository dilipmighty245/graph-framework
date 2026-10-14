package parser

import (
	"go/ast"
	goparser "go/parser"
	"testing"
)

func TestNodeResolutionUsesImportQualifier(t *testing.T) {
	nodes := map[string]Node{
		"first":  {Name: "User", PkgName: "first", FullName: "example.test/model/first", CrdName: "users.first.example.test"},
		"second": {Name: "User", PkgName: "second", FullName: "example.test/model/second", CrdName: "users.second.example.test"},
	}
	base := &Node{Imports: []*ast.ImportSpec{
		{Path: &ast.BasicLit{Value: `"example.test/model/first"`}},
		{Name: ast.NewIdent("alias"), Path: &ast.BasicLit{Value: `"example.test/model/second"`}},
	}}
	for expression, expected := range map[string]string{"first.User": "users.first.example.test", "alias.User": "users.second.example.test"} {
		typ, err := goparser.ParseExpr(expression)
		if err != nil {
			t.Fatal(err)
		}
		found := findNodeDefForField(&ast.Field{Type: typ}, base, nodes)
		if found == nil || found.CrdName != expected {
			t.Fatalf("%s resolved to %#v; want %s", expression, found, expected)
		}
	}
}
