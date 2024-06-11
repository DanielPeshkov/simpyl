## Welcome to the Simplified Python Language (Simpyl)!
Simpyl is a dynamically-typed interpreted programming language with Python-like sytax. In fact, Simpyl syntax is a subset of Python, meaning valid Simpyl code is valid Python code. The language is written in Go, and the current architecture includes a Pratt parser and Tree-Walking Interpreter. The current implementation supports Integer, Float, Boolean and String data types, List, Dictionary and Set data structures, For and While loops, Functions, and a variety of useful builtin functions. Going forward, I plan to create a bytecode compiler and virtual machine for Simpyl and further optimize the language to improve performance. 

#### Benchmark Results
In order to guage the speed of this language in comparison to other programming languages, I have included two benchmark functions: the leibniz formula for pi and the recursive fibonacci function. 

The leibniz function calculates the Leibniz formula for pi 100 million times, and the execution time was recorded for Simpyl, Python and Go. These are the results:
- Simpyl v0.1.0  :: 21.25s
- Python v3.10.9 :: 5.125s
- Go v1.22.3     :: 0.5s

The fibonacci function calculates the 36th number in the fibonacci sequence, and the execution time was recorded for Simpyl, Python and Go. These are the results:
Fibonacci(36)
- Simpyl v0.1.0 :: 16.2s
- Python v3.10.9 :: 2.45s
- Go v1.22.3 :: 0.57s