# hypescript

HypeScript is a TypeScript compiler that produces native binaries for target platforms.

It compiles to standardized C++ so as long as your platform has a standards-compliant C++ compiler, you should be good to go!

## TODO

- `function`s
    - [x] Named
    - [x] Anonymous
    - [ ] Hoisting
- Objects
    - [x] literals
    - [x] chained access
    - [x] invocation
    - [ ] methods
- [x] Closures
- Variables
    - [x] `let` declarations
    - [ ] `const` declarations
    - [ ] `var` declarations
    - [ ] allow value-less bindings
- Types
    - [x] `function` params
    - [ ] variable types
    - `interface`s
        - [ ] declarations
        - [ ] type checking
    - `class`es
        - [ ] declarations
        - [ ] type checking
    - [x] Mutually recursive types
- [x] Scoped identifiers
- [ ] `this`
- [ ] `import`s
- [ ] `export`s
- [ ] modules
- [ ] declaration files