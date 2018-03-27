Ontology wasm
=====
## Introduction
Ontology wasm is a VM for ontology block chain, it can also be used for other stand-alone environment not only for block chains.

WebAssembly (abbreviated *Wasm*) is a binary instruction format for a stack-based virtual machine. Wasm is designed as a portable target for compilation of high-level languages like C/C++/Rust.



## Structure

Ontology-wasm disassemble wasm binary codes based on [Wagon](https://github.com/go-interpreter/wagon) project with extra memory management. 

Currently, we support int, int64 float, double, string(byte array), int array and int64 array data types,

since wasm only has 4 types (i32,i64,f32 and f64), that means only the 4 types data could be pushed into the stack,any other complex data type must be stored in memory.

In  MVP version,every module can has only one linear memory

![memory](./doc/images/memory.png)

