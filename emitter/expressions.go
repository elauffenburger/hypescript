package emitter

import (
	"elauffenburger/hypescript/ast"
	"fmt"
	"strings"
)

func writeStatementOrExpression(ctx *Context, stmtOrExpr *ast.StatementOrExpression) error {
	if stmtOrExpr.Statement != nil {
		return writeStatement(ctx, stmtOrExpr.Statement)
	}

	if stmtOrExpr.Expression != nil {
		return writeExpression(ctx, stmtOrExpr.Expression)
	}

	return fmt.Errorf("unknown StatementOrExpression: %#v", stmtOrExpr)
}

func writeStatement(ctx *Context, stmt *ast.Statement) error {
	if exprStmt := stmt.ExpressionStmt; exprStmt != nil {
		err := writeExpression(ctx, exprStmt)
		if err != nil {
			return err
		}

		ctx.WriteString(";")

		return nil
	}

	if letDecl := stmt.LetDecl; letDecl != nil {
		letDeclType, err := inferType(ctx, &letDecl.Value)
		if err != nil {
			return err
		}

		var typeName string
		if nonUnionType := letDeclType.NonUnionType; nonUnionType != nil {
			if nonUnionType.TypeReference != nil {
				typeName = mangleTypeName(*letDeclType.NonUnionType.TypeReference)
			} else if literalType := nonUnionType.LiteralType; literalType != nil {

				if literalType.ObjectType != nil {
					typeName = "ts_object*"
				}
			}
		}

		ctx.WriteString(fmt.Sprintf("%s %s = ", typeName, letDecl.Name))

		err = writeExpression(ctx, &letDecl.Value)
		if err != nil {
			return err
		}

		ctx.WriteString(";")

		ctx.CurrentScope.AddIdentifer(letDecl.Name, *letDeclType)

		return nil
	}

	if assignmentStmt := stmt.AssignmentStmt; assignmentStmt != nil {
		ctx.WriteString(fmt.Sprintf("%s = ", assignmentStmt.Ident))

		err := writeExpression(ctx, &assignmentStmt.Value)
		if err != nil {
			return err
		}

		ctx.WriteString(";")

		return nil
	}

	if returnStmt := stmt.ReturnStmt; returnStmt != nil {
		ctx.WriteString("return ")

		err := writeExpression(ctx, returnStmt)
		if err != nil {
			return err
		}

		ctx.WriteString(";")

		return nil
	}

	return fmt.Errorf("unknown statement type: %#v", stmt)
}

func writeExpression(ctx *Context, expr *ast.Expression) error {
	if expr.Number != nil {
		if expr.Number.Integer != nil {
			ctx.WriteString(fmt.Sprintf("ts_num_new(%d)", *expr.Number.Integer))
			return nil
		}

		return fmt.Errorf("unknown number expression: %#v", *expr)
	}

	if expr.String != nil {
		ctx.WriteString(fmt.Sprintf("ts_string_new(%s)", *expr.String))
		return nil
	}

	if objInst := expr.ObjectInstantiation; objInst != nil {
		return writeObjectInstantiation(ctx, objInst)
	}

	if chainedObjOperation := expr.ChainedObjectOperation; chainedObjOperation != nil {
		// TODO: need to update this as we walk the chain.
		accessee := &chainedObjOperation.Accessee

		for _, operation := range chainedObjOperation.Operations {
			if objAccess := operation.Access; objAccess != nil {
				nextAccessee, err := writeAccessedObjectAccess(ctx, accessee, operation.Access)
				if err != nil {
					return err
				}

				accessee = nextAccessee

				continue
			}

			if objInvoc := operation.Invocation; objInvoc != nil {
				nextAccessee, err := writeAccessedObjectInvocation(ctx, accessee, operation.Invocation)
				if err != nil {
					return err
				}

				accessee = nextAccessee

				continue
			}

			return fmt.Errorf("unknown operation in chained object operation: %#v", operation)
		}

		return nil
	}

	if expr.Ident != nil {
		return writeIdent(ctx, *expr.Ident)
	}

	return fmt.Errorf("unknown expression type: %#v", expr)
}

func writeAccessedObjectAccess(ctx *Context, accessee *ast.Accessable, access *ast.ObjectAccess) (nextAccessee *ast.Accessable, err error) {
	// TODO: this will _not_ work for multiple chained object accessess
	// because of how ts_object_get_field calls will be emitted!

	accesseeType, err := inferAccessableType(ctx, *accessee)
	if err != nil {
		return nil, err
	}

	// Figure out how to access the accessee.
	if accesseeType.NonUnionType != nil {
		if accesseeType.NonUnionType.LiteralType != nil {
			// Check if this is an object type.
			if accesseeType.NonUnionType.LiteralType.ObjectType != nil {
				fieldName := access.AccessedIdent
				var field *ast.ObjectTypeField

				for _, f := range accesseeType.NonUnionType.LiteralType.ObjectType.Fields {
					if f.Name == fieldName {
						field = &f

						break
					}
				}

				if field == nil {
					return nil, fmt.Errorf("failed to find field %s in %#v", fieldName, accesseeType)
				}

				fieldType, err := getRuntimeTypeName(&field.Type)
				if err != nil {
					return nil, err
				}

				ctx.WriteString(fmt.Sprintf("(%s*)ts_object_get_field(", mangleTypeName(fieldType)))

				err = writeAccessableAccess(ctx, *accessee)
				if err != nil {
					return nil, err
				}

				ctx.WriteString(fmt.Sprintf(", \"%s\")", field.Name))

				return typeToAccessee(&field.Type)
			}
		}
	}

	return nil, fmt.Errorf("unknown type in object access: %#v", accesseeType)
}

func typeToAccessee(t *ast.Type) (*ast.Accessable, error) {
	if t := t.NonUnionType; t != nil {
		if t := t.LiteralType; t != nil {
			return &ast.Accessable{LiteralType: t}, nil
		}

		if t := t.TypeReference; t != nil {
			return &ast.Accessable{Ident: t}, nil
		}
	}

	return nil, fmt.Errorf("could not convert type to accessee: %#v", t)
}

func getRuntimeTypeName(t *ast.Type) (string, error) {
	if t.NonUnionType != nil {
		if t.NonUnionType.TypeReference != nil {
			return *t.NonUnionType.TypeReference, nil
		}

		if t.NonUnionType.LiteralType != nil {
			if t.NonUnionType.LiteralType.ObjectType != nil {
				return string(TsObject), nil
			}
		}
	}

	return "", fmt.Errorf("unknown type: %#v", t)
}

func writeAccessedObjectInvocation(ctx *Context, accessee *ast.Accessable, invoc *ast.ObjectInvocation) (nextAccessee *ast.Accessable, err error) {
	accesseeType, err := inferAccessableType(ctx, *accessee)
	if err != nil {
		return nil, err
	}

	// Figure out how to access the accessee.
	if t := accesseeType.NonUnionType; t != nil {
		if t := t.LiteralType; t != nil {
			if t.ObjectType != nil {
				ctx.WriteString(*accessee.Ident)

				err := writeObjectInvocation(ctx, invoc)
				if err != nil {
					return nil, err
				}
			}

			if t.FunctionType != nil {
				ctx.WriteString(*accessee.Ident)

				err := writeObjectInvocation(ctx, invoc)
				if err != nil {
					return nil, err
				}
			}

			return typeToAccessee(accesseeType)
		}
	}

	return nil, fmt.Errorf("unknown type in object invocation: %#v", accesseeType)
}

func writeObjectInvocation(ctx *Context, invocation *ast.ObjectInvocation) error {
	// TODO: support arguments.
	ctx.WriteString("()")

	return nil
}

func inferAccessableType(ctx *Context, accessable ast.Accessable) (*ast.Type, error) {
	if accessable.Ident != nil {
		return ctx.TypeOf(*accessable.Ident)
	}

	if accessable.LiteralType != nil {
		return &ast.Type{NonUnionType: &ast.NonUnionType{LiteralType: accessable.LiteralType}}, nil
	}

	return nil, fmt.Errorf("unknown accessable type: %#v", accessable)
}

func writeObjectInstantiation(ctx *Context, objInst *ast.ObjectInstantiation) error {
	formattedFields := strings.Builder{}
	for i, field := range objInst.Fields {
		name := field.Name

		fieldType, err := inferType(ctx, &field.Value)
		if err != nil {
			return err
		}

		typeId, err := getTypeIdFor(ctx, fieldType)
		if err != nil {
			return err
		}

		// TODO: handle actual cases of types that need metadata.
		metadata := "NULL"

		fieldDescriptor := fmt.Sprintf("ts_object_field_descriptor_new(ts_string_new(\"%s\"), %d, %s)", name, typeId, metadata)

		value, err := ctx.WithinPrintContext(func(printCtx *Context) error {
			return writeExpression(printCtx, &field.Value)
		})

		if err != nil {
			return err
		}

		formattedFields.WriteString(fmt.Sprintf("ts_object_field_new(%s, %s)", fieldDescriptor, value))

		if i != len(objInst.Fields)-1 {
			formattedFields.WriteString(", ")
		}
	}

	ctx.WriteString(fmt.Sprintf("ts_object_new((ts_object_field*[]){%s})", formattedFields.String()))

	return nil
}

func writeIdent(ctx *Context, ident string) error {
	identType, err := ctx.TypeOf(ident)

	// If we couldn't find the type of the ident, don't try mangling it --
	// it's either not defined yet or it's a bug we'll catch later.
	if err != nil {
		ctx.WriteString(ident)
	} else {
		// Otherwise, mangle away!
		ctx.WriteString(mangleIdentName(ident, identType))
	}

	return nil
}

func writeAccessableAccess(ctx *Context, accessable ast.Accessable) error {
	if accessable.Ident != nil {
		ctx.WriteString(*accessable.Ident)

		return nil
	}

	if accessable.LiteralType != nil {
		return nil
	}

	return fmt.Errorf("unknown accessable type: %#v", accessable)
}

func getTypeIdFor(ctx *Context, t *ast.Type) (int, error) {
	// DO NOT SUBMIT; need to actually make this work!

	return 0, nil
}
