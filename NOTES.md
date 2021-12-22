# Notes

## Functions

### Hoisting

I can probably forward-declare the `TsFunction*` variables by looking for named `function` declarations in the current scope, and then assign them as I'm actually writing out the `function`s.

## Mutual Recursion

We need to support allowing constructs like `function`s, `class`es, `interface`s, and `type`s (maybe others?) to used before they're actually defined.

This happens in situations like:

```ts
interface Foo {
    bar: Bar;
}

interface Bar {
    baz: string;
}
```

Right now, we don't allow this because we're just going to complain that we don't know about `Bar` when declaring `Foo`.

I suspect we'll need to do multiple passes: one to create a symbol table of everything at each scope, and then a second to actually do codegen based on that symbol table.

This is going to require a rework of how we currently do things but it's almost certainly necessary. I'm thinking I'll get to a good place with unit tests, etc. and then make the switch to make sure that functionality continues to work as expected.

On looking at this, I wonder if we can get away with _just_ recording the things that would normally need to be forward-declared? We basically already do this for `function` bodies: we build `functionInfo` structs that figure out the implicit return type to make sure it matches the explicit one. So maybe we need to just record the `function`s, `class`es, `interface`s, and `type`s (and whatever else) in each scope, then come back once we've resolved all types and emit the code (at which point we'll _know_ if an identifier is actually missing).