package main

import (
	"reflect"
	"testing"

	"github.com/alecthomas/participle"
)

func TestParseSimpleFunction(t *testing.T) {
	ast := parseString(
		`
		function foo (bar: string, baz: num): num {
			let foo = 5;
			let bar = "bar";

			return foo;
		}
	`, t)

	assertEqual(t, ast, &TS{
		Functions: []Function{
			{
				Name: "foo",
				ReturnType: &Type{
					TypeName: "num",
				},
				Arguments: []FunctionArgument{
					{Name: "bar", Type: Type{TypeName: "string"}},
					{Name: "baz", Type: Type{TypeName: "num"}},
				},
				Body: []ExpressionOrStatement{
					{
						Statement: &Statement{
							LetDecl: &LetDecl{
								Name: "foo",
								Value: Expression{
									Number: &Number{
										Integer: 5,
									},
								},
							},
						},
					},
					{
						Statement: &Statement{
							LetDecl: &LetDecl{
								Name: "bar",
								Value: Expression{
									String: "bar",
								},
							},
						},
					},
					{
						Statement: &Statement{
							ReturnStmt: &Expression{Ident: "foo"},
						},
					},
				},
			},
		},
	})
}

func assertEqual(t *testing.T, actual *TS, expected *TS) {
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("actual did not match expected: \n\nACTUAL:\n%#v\n\nEXPECTED:\n%#v", actual, expected)
	}
}

func parseString(str string, t *testing.T) *TS {
	parser, err := participle.Build(&TS{})
	if err != nil {
		t.Error(err)
	}

	ast := TS{}
	err = parser.ParseString(str, &ast)

	if err != nil {
		t.Error(err)
	}

	return &ast
}
