package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"simpyl/evaluator"
	"simpyl/lexer"
	"simpyl/object"
	"simpyl/parser"
	"strings"
)

const PROMPT = ">> "

func StartInteractive(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}

func StartInterpreter(file string) {
	f, err := os.ReadFile(file)
	if err != nil {
		fmt.Print(err)
	}

	str := string(f)
	str = strings.Replace(str, "    ", "\t", -1)
	l := lexer.New(str)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		panic(strings.Join(p.Errors(), "\n"))
	}

	env := object.NewEnvironment()
	evaluated := evaluator.Eval(program, env)
	if evaluated != nil && evaluated != evaluator.NULL {
		println(evaluated.Inspect())
	}
}
