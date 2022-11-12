function fizzbuzz(n: number): void {
  for (let i = 1; i < n + 1; i++) {
    let fizz: boolean = i % 3 == 0;
    let buzz: boolean = i % 5 == 0;

    if (fizz && buzz) {
      console.log("fizzbuzz");
    } else if (fizz) {
      console.log("fizz");
    } else if (buzz) {
      console.log("buzz");
    } else {
      console.log(i);
    }
  }
}

fizzbuzz(10);
