package writepass

import (
	"elauffenburger/hypescript/emitter/core"
	"fmt"
	"strings"
)

func (ctx *Context) writeFunctionDeclaration(fn *core.Function) error {
	ctx.WriteString(fmt.Sprintf("TsFunction* %s = ", mangleFunctionName(*fn.Name)))

	err := ctx.writeFunction(fn)
	if err != nil {
		return err
	}

	ctx.WriteString(";")

	return nil
}

func (ctx *Context) writeFunction(fn *core.Function) error {
	// Format the function params.
	formattedParams := strings.Builder{}
	formattedParams.WriteString("TsCoreHelpers::toVector<TsFunctionParam>({")

	// Figure out the function name.
	var fnName string
	if fn.Name != nil {
		fnName = *fn.Name
	} else {
		fnName = ctx.currentScope().NewIdent()
	}

	numParams := len(fn.Parameters)
	for i, p := range fn.Parameters {
		typeId, err := ctx.currentScope().GetTypeIdFor(p.Type)
		if err != nil {
			return err
		}

		formattedParams.WriteString(fmt.Sprintf("TsFunctionParam(\"%s\", %d)", p.Name, typeId))

		if i != numParams-1 {
			formattedParams.WriteString(", ")
		}
	}

	formattedParams.WriteString("})")

	ctx.WriteString(
		fmt.Sprintf(
			"new TsFunction(\"%s\", %s, ",
			fnName,
			formattedParams.String(),
		),
	)

	err := ctx.writeFunctionLambda(fn)
	if err != nil {
		return err
	}

	ctx.WriteString(")")

	return nil
}

func (ctx *Context) writeFunctionLambda(fn *core.Function) error {
	ctx.WriteString("[=](TsObject* _this, std::vector<TsFunctionArg> args) -> TsObject* {")

	// Unpack each arg into local vars in the function.
	for _, param := range fn.Parameters {
		ctx.WriteString(fmt.Sprintf("auto %s = TsFunctionArg::findArg(args, \"%s\").value;", param.Name, param.Name))
	}

	// Write the body.
	for _, stmtOrExpr := range fn.Body {
		err := ctx.writeStatementOrExpression(stmtOrExpr)
		if err != nil {
			return err
		}
	}

	if t := fn.ImplicitReturnType.TypeReference; t != nil && *t == string(core.RtTsVoid) {
		ctx.WriteString("return NULL;")
	}

	ctx.WriteString("}")

	return nil
}
