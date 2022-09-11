# hypescript

HypeScript is a TypeScript compiler that produces native binaries for target platforms.

It compiles to Rust, so as long as your platform has a Rust toolchain installed, you should be good to go!

## TODO

- `function`s
    - [x] named
    - [x] anonymous
    - [ ] hoisting
    - [x] closures
- Objects
    - [x] literals
    - [x] chained access
    - [x] invocation
    - [x] methods
- [x] Closures
- Variables
    - [ ] `let` declarations
    - [x] `const` declarations
    - [ ] `var` declarations
- Types
    - [x] `function` params
        - [ ] optional params
        - [x] typechecking
    - [ ] variable types
    - `interface`s
        - [x] declarations
        - [x] type checking
    - `class`es
        - [ ] declarations
        - [ ] type checking
    - `type`s
    - [ ] mutually recursive types
    - [ ] `any`
    - [ ] `undefined`
    - [ ] `null`
    - unions
        - [ ] variables
        - [ ] type checking
    - [ ] generics
    - [ ] arrays
- [ ] scoped identifiers
- [x] `this`
- [ ] `new`
- [ ] `import`s
- [ ] `export`s
- [ ] modules
- [ ] declaration files