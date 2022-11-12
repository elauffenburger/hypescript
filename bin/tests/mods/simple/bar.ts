import { Foo } from "./foo";

interface Bar {
  foo: Foo;
}

let bar: Bar = {
  foo: {
    name: "hi i'm a foo!",
  },
};

console.log(bar.foo.name);
