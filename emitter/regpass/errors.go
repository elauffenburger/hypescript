package regpass

import (
	"elauffenburger/hypescript/emitter/core"
	"fmt"
)

type TypeMismatchError struct {
	Name     string
	Expected *core.TypeSpec
	Actual   *core.TypeSpec
}

func (err TypeMismatchError) Error() string {
	return fmt.Sprintf(
		"%s had a type annotation of %#v, but the type was found to be %#v",
		err.Name,
		err.Expected,
		err.Actual,
	)
}

type FnRtnTypeMismatchError struct {
	Name     string
	Implicit *core.TypeSpec
	Explicit *core.TypeSpec
}

func (err FnRtnTypeMismatchError) Error() string {
	return fmt.Sprintf(
		"%s had an explicit return type of %#v, but the implicit return type was found to be %#v",
		err.Name,
		err.Explicit,
		err.Implicit,
	)
}
