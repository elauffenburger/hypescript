# Notes

## Functions

### Hoisting

I can probably forward-declare the `TsFunction*` variables by looking for named `function` declarations in the current scope, and then assign them as I'm actually writing out the `function`s.

This may just be how I need to approach declarations in general; probably need to switch to a multi-pass approach [link](https://en.wikipedia.org/wiki/Multi-pass_compiler).

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