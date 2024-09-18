# chlang
The Chlang Programming Language

**WARNING: This language is under development. Most features not implemented yet and can be changed in the feature. Use it at your own risk.**

The main goal of this project is to create a statically and strong typed programming language. The language is inspired by Lua, Kotlin and Rust. It uses a register-based virtual machine like [Lua](https://the-ravi-programming-language.readthedocs.io/en/latest/lua_bytecode_reference.html)

The register-based code can be used as an intermediate representation to generate code for native platforms like x86, ARM, etc.

## Planned features:
- [x] Variable, constant and function declarations
- [x] Type checking and inference
- [x] Function calls
- [x] Control structures (if, while, for)
- [ ] Modules and packages
- [ ] Data structures (arrays, vectors, user-defined structures)
- [x] Interpreter for register-based virtual machine

## Example
```rust
package main

const PI = 3.14159

struct Blog {
    title: string,
    content: string,
    likes: i32,
}

fn circle_area(radius: i32) -> f64 {
    return f64(PI) * radius * radius
}

fn multi_return(a: i32) -> i32, i32, f64 {
    return a, 2, 2.65
}

fn fibonacci_linear(n: i32) -> i32 { 
    let a = 0
    let b = 1
    for i in 0..=n {
        let c = a + b
        a, b = b, c
    }
    return a
}

fn main() {
    let radius = 10
    let area = circle_area(radius)
    let inside = if area > 100 {
        println("Area is greater than 100")
        true
    } else {
        println("Area is less than 100")
        false
    }
    println("Area: $area ($inside)")
    println("Fibonacci(10): ${fibonacci_linear(10)}")
    
    let a, b, c = multi_return(10)
    println("Multi return: $a, $b, $c")
}
```
