package writepass

import (
	"elauffenburger/hypescript/emitter/core"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

func validate(f *core.Function) error {
	// Make sure the implicit return type matches the implicit one (if any).
	if rtnType := f.ExplicitReturnType; rtnType != nil {
		if !rtnType.EqualsReferencing(f.ImplicitReturnType) {
			return errors.WithStack(fmt.Errorf("implicit and explicit return types of function were not the same: %#v", *f))
		}
	}

	return nil
}

func (ctx *Context) writeFunctionDeclaration(fn *core.Function) error {
	if err := validate(fn); err != nil {
		return err
	}

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

	var fnName string
	if fn.Name != nil {
		fnName = *fn.Name
	} else {
		fnName = ctx.currentScope().NewIdent()
	}

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
	ctx.WriteString("[=](std::vector<TsFunctionArg> args) -> TsObject* {")

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
