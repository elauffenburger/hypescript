package emitter

import (
	"bufio"
	"elauffenburger/hypescript/ast"
	_ "embed"
	"io"
)

//go:embed runtime.txt
var runtime string

type primitiveType string

const (
	TsString primitiveType = "string"
	TsNum    primitiveType = "num"
)

type coreType string

const (
	TsObject coreType = "ts_object"
	TsVoid   coreType = "void"
)

type Emitter interface {
	Emit(ast *ast.TS) error
}

type emitter struct {
	writer io.Writer
}

func (e emitter) Emit(ast *ast.TS) error {
	context := NewContext(bufio.NewWriter(e.writer))

	writeRuntime(context)

	for _, function := range ast.Functions {
		err := writeFunction(context, &function)
		if err != nil {
			return err
		}

		context.WriteString("\n\n")
	}

	return context.Output.Flush()
}

func New(writer io.Writer) Emitter {
	return &emitter{writer: writer}
}

func writeRuntime(ctx *Context) {
	ctx.WriteString(runtime)
}
