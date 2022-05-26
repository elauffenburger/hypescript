package regpass

import (
	"elauffenburger/hypescript/ast"
	"elauffenburger/hypescript/emitter/core"
	"fmt"
)

func (ctx *Context) chainedObjOperationFromAst(chainedOp *ast.ChainedObjectOperation) (*core.ChainedObjectOperation, error) {
	var firstLink *core.ObjectOperation
	var currentLink *core.ObjectOperation

	accessee, err := ctx.accessableFromAst(chainedOp.Accessee)
	if err != nil {
		return nil, err
	}

	for _, astOp := range chainedOp.Operations {
		// Create a new link.
		link := &core.ObjectOperation{
			Accessee: accessee,
			Prev:     currentLink,
		}

		if firstLink == nil {
			firstLink = link
		}

		// Link the current link (if any) to the new link.
		if currentLink != nil {
			currentLink.Next = link
		}

		// Make the current link the new link.
		currentLink = link

		// Grab some parent info if available.
		var parent core.HasMembers
		{
			parentType, err := ctx.currentScope().ResolveType(accessee.Type)
			if err != nil {
				return nil, err
			}

			if parentType.Object != nil || parentType.Interface != nil {
				if parentType.Interface != nil {
					parent = parentType.Interface
				} else if parentType.Object != nil {
					parent = parentType.Object
				}
			}
		}

		// Add "access" chain operation.
		if access := astOp.Access; access != nil {
			ident := access.AccessedIdent
			var identType *core.TypeSpec
			{
				member, ok := parent.AllMembers()[ident]
				if ok {
					identType = member.Type()
				} else {
					c := make(chan *core.Member)
					go func() {
						c <- parent.WaitForMember(ident)
					}()

					select {
					case <-ctx.Done():
					case member := <-c:
						identType = member.Type()
					}
				}
			}

			link.Access = &core.ObjectAccess{AccessedIdent: ident, Type: identType}
			accessee = &core.Accessable{Ident: &ident, Type: identType}

			continue
		}

		// Add "invocation" chain operation.
		if astInvoc := astOp.Invocation; astInvoc != nil {
			invoc, err := ctx.invocationFromAst(accessee, astInvoc)
			if err != nil {
				return nil, err
			}

			link.Invocation = invoc

			if t := accessee.Type; t != nil {
				if fn := accessee.Type.Function; fn != nil {
					accessee = &core.Accessable{Ident: fn.Name, Type: fn.ImplicitReturnType}

					continue
				}

				return nil, fmt.Errorf("could not resolve return type of %#v", accessee)
			}

			continue
		}

		return nil, fmt.Errorf("unknown operation in chained object operation: %#v", astOp)
	}

	// Check if there's an assignment.
	if assign := chainedOp.Assignment; assign != nil {
		expr, err := ctx.expressionFromAst(&assign.Value)
		if err != nil {
			return nil, err
		}

		link := &core.ObjectOperation{
			Accessee: accessee,
			Assignment: &core.Assignment{
				Value: expr,
			},
		}

		// This is a little weird, but we want to remove the
		// previous link in the case of an assignment because
		// the last link would have _looked_ like an access, but
		// it's actually an assignment.
		//
		// i.e. `[access foo]->[access bar on foo]->[assign bar on bar "baz"]` becomes:
		//      `[access foo]->[assign bar on foo "baz"]`
		if currentLink != nil {
			link.Prev = currentLink.Prev
			currentLink.Prev.Next = link
		}

		currentLink = link
	}

	return &core.ChainedObjectOperation{First: firstLink, Last: currentLink}, nil
}

func (ctx *Context) invocationFromAst(accessee *core.Accessable, invoc *ast.ObjectInvocation) (*core.ObjectInvocation, error) {
	args := make([]*core.Expression, len(invoc.Arguments))
	for i, a := range invoc.Arguments {
		expr, err := ctx.expressionFromAst(&a)
		if err != nil {
			return nil, err
		}

		args[i] = expr
	}

	return &core.ObjectInvocation{Accessee: accessee, Arguments: args}, nil
}
