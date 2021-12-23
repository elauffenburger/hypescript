package writepass

import "elauffenburger/hypescript/emitter/core"

func mangleFunctionName(name string) string {
	return name
}

func mangleIdentName(name string, identType *core.TypeSpec) string {
	if identType.Function != nil {
		return mangleFunctionName(name)
	}

	return name
}
