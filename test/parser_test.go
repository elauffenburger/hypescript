package main

import (
	"elauffenburger/hypescript/ast"
	"reflect"
	"testing"

	"github.com/alecthomas/participle"
)

func TestParseSimpleFunction(t *testing.T) {
	parsed := parseString(
		`
		function foo (bar: string, baz: num): num {
			let foo = 5;
			let bar = "bar";

			return foo;
		}
	`, t)

	assertEqual(t, parsed, &ast.TS{
		TopLevelConstructs: []ast.TopLevelConstruct{
			{
				StatementOrExpression: &ast.StatementOrExpression{
					Statement: &ast.Statement{
						FunctionInstantiation: &ast.FunctionInstantiation{
							Name: strRef("foo"),
							ReturnType: &ast.TypeIdentifier{
								NonUnionType: &ast.NonUnionType{TypeReference: strRef("num")},
							},
							Parameters: []*ast.FunctionParameter{
								{
									Name: "bar",
									Type: ast.TypeIdentifier{
										NonUnionType: &ast.NonUnionType{TypeReference: strRef("string")},
									},
								},
								{
									Name: "baz",
									Type: ast.TypeIdentifier{
										NonUnionType: &ast.NonUnionType{TypeReference: strRef("num")},
									},
								},
							},
							Body: []*ast.StatementOrExpression{
								{
									Statement: &ast.Statement{
										LetDecl: &ast.LetDecl{
											Name: "foo",
											Value: ast.Expression{
												Number: &ast.Number{Integer: intRef(5)},
											},
										},
									},
								},
								{
									Statement: &ast.Statement{
										LetDecl: &ast.LetDecl{
											Name:  "bar",
											Value: ast.Expression{String: strRef("bar")},
										},
									},
								},
								{
									Statement: &ast.Statement{
										ReturnStmt: &ast.Expression{Ident: strRef("foo")},
									},
								},
							},
						},
					},
				},
			},
		},
	})
}

func strRef(str string) *string {
	return &str
}

func intRef(num int) *int {
	return &num
}

func assertEqual(t *testing.T, actual *ast.TS, expected *ast.TS) {
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("actual did not match expected: \n\nACTUAL:\n%#v\n\nEXPECTED:\n%#v", actual, expected)
	}
}

func parseString(str string, t *testing.T) *ast.TS {
	parser, err := participle.Build(&ast.TS{})
	if err != nil {
		t.Error(err)
	}

	ast := ast.TS{}
	err = parser.ParseString(str, &ast)

	if err != nil {
		t.Error(err)
	}

	return &ast
}
