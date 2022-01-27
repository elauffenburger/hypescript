package writepass

import "elauffenburger/hypescript/emitter/core"

func mangleFunctionName(name string) string {
	return name
}

func mangleIdentName(name string, identType *core.TypeSpec) string {
	if identType.Function != nil {
		return mangleFunctionName(name)
	}

	if name == string(core.TsThis) {
		// TODO: this won't always work.
		return "_this"
	}

	return name
}
