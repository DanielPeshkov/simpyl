package evaluator

import (
	"simpyl/lexer"
	"simpyl/object"
	"simpyl/parser"
	"testing"
)

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

/*
Statement Testing
*/

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{
			`
if 10 > 1:
	if 10 > 1:
		return 10;
	return 1;
`,
			10,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

/*
Literal Testing
*/
func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalFloatExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"5.", 5},
		{"10.", 10},
		{"-.5", -.5},
		{"-.1", -.1},
		{".5 + .5 + .5 + .5 - 10.", -8},
		{"2 * 2 * 2 * 2 * .2", 3.2},
		{"-.50 + 100. + -.50", 99},
		{".5 * 2 + 10", 11},
		{"5 + 2 * 1.0", 7},
		{"20. + .2 * -10", 18},
		{"50 / 2 * .2 + 10", 15},
		{"2 * (5 + 1.0)", 12},
		{"3 * 3. * 3 * .10", 2.7},
		{"3 * (.3 * .3) + 1.0", 1.27},
		{"(5 + 10 * .2 + 15 / 3) * 2 + -10", 14},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello World!"`
	evaluated := testEval(input)

	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestListLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.List)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d",
			len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestListIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			3,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestIndexAssignExpressions(t *testing.T) {
	input := `arr = [0, 1, 2]
arr[0] = 3
return arr[0]`
	evaluated := testEval(input)

	testIntegerObject(t, evaluated, 3)
}

func TestDictLiterals(t *testing.T) {
	input := `two = "two"
{
	"one": 10 - 9,
	two: 1 + 1,
	"thr" + "ee": 6 / 2,
	4: 4,
	true: 5,
	false: 6}`
	evaluated := testEval(input)

	result, ok := evaluated.(*object.Dict)
	if !ok {
		t.Fatalf("Eval didn't return Dict. got=%T (%+v)", evaluated, evaluated)
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("Dict has wrong num of pairs. got=%d", len(result.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}

		testIntegerObject(t, pair.Value, expectedValue)
	}
}

func TestDictIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`let key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{true: 5}[true]`,
			5,
		},
		{
			`{false: 5}[false]`,
			5,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`
if true:
	return 10`,
			10},
		{`
if false:
	return 10`,
			nil},
		{`
if 1:
	return 10`,
			10},
		{`
if 1 < 2:
	return 10`,
			10},
		{`
if 1 > 2:
	return 10`,
			nil},
		{`
if 1 > 2:
	return 10
else:
	return 20`,
			20},
		{`
if 1 < 2:
	a = 10
else:
	a = 20
a`,
			10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestInExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`
s = set()
s.add(1)
return 1 in s`,
			true},
		{`
s = set()
s.add(1)
return 2 in s`,
			false},
		{`
l = [1, 2, 3]
return 2 in l`,
			true},
		{`
d = {"a": 1, "b": 2}
return "a" in d.keys()`,
			true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		testBooleanObject(t, evaluated, tt.expected)

	}
}

/*
Function Testing
*/
func TestFunctionObject(t *testing.T) {
	input := `
def func(x):
	x + 2
func`
	evaluated := testEval(input)

	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`
def func(x):
	return x+1
return func(5)`, 6},
		{`
def func(x):
	return x
func(5)`, 5},
		{`
def double(x):
	return 2 * x
double(5)`, 10},
		{`
def add(x, y):
	return x + y
add(5, 5)`, 10},
		{`
def add(x, y):
	return x + y
add(5+5, add(5, 5))`, 20},
		{`
def f(x):
	x
f(5)`, 5},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
def newAdder(x):
	def f(y):
		return x + y
	return f
	
addTwo = newAdder(2)
addTwo(2)`
	testIntegerObject(t, testEval(input), 4)
}

func TestObjectMethod(t *testing.T) {
	input := `
list = []
list.append(5)
list[0]`
	testIntegerObject(t, testEval(input), 5)
}

/*
Loop Testing
*/
func TestForLoop(t *testing.T) {
	input := `x = 0
for i in range(5):
	x = x + i
return x`
	evaluated := testEval(input)

	testIntegerObject(t, evaluated, 10)
}

func TestWhileLoop(t *testing.T) {
	input := `x = 0
while x < 5:
	x = x + 1
return x`
	evaluated := testEval(input)

	testIntegerObject(t, evaluated, 5)
}

func TestForInFunctionStatement(t *testing.T) {
	input := `
def foo(x):
	for i in range(5):
		x = x + i
	return x

x = foo(0)
return x`
	evaluated := testEval(input)

	testIntegerObject(t, evaluated, 10)
}

/*
Builtin Function Testing
*/
func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)",
					evaluated, evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q",
					expected, errObj.Message)
			}
		}
	}
}

func TestBuiltinRange(t *testing.T) {
	input := `range(1+2)`
	evaluated := testEval(input)

	list, ok := evaluated.(*object.List)
	if !ok {
		newError("range did not return list, got=%T", evaluated.Type())
	}
	e := list.Elements

	if len(e) != 3 {
		newError("range returned incorrect list size. expected=%d elements, got=%d", 3, len(e))
	}

	testIntegerObject(t, e[0], 0)
	testIntegerObject(t, e[1], 1)
	testIntegerObject(t, e[2], 2)
}

func TestBuiltinMin(t *testing.T) {
	input := `min([1, 2, 3])`
	evaluated := testEval(input)

	i, ok := evaluated.(*object.Integer)
	if !ok {
		newError("min did not return Integer, got=%T", evaluated.Type())
	}

	testIntegerObject(t, i, 1)
}

func TestBuiltinMax(t *testing.T) {
	input := `max([1, 2, 3])`
	evaluated := testEval(input)

	i, ok := evaluated.(*object.Integer)
	if !ok {
		newError("max did not return Integer, got=%T", evaluated.Type())
	}

	testIntegerObject(t, i, 3)
}

func TestBuiltinAbs(t *testing.T) {
	input := `abs(-1)`
	evaluated := testEval(input)

	i, ok := evaluated.(*object.Integer)
	if !ok {
		newError("abs did not return Integer, got=%T", evaluated.Type())
	}

	testIntegerObject(t, i, 1)
}

func TestBuiltinSum(t *testing.T) {
	input := `sum([1, 2, 3])`
	evaluated := testEval(input)

	i, ok := evaluated.(*object.Integer)
	if !ok {
		newError("sum did not return Integer, got=%T", evaluated.Type())
	}

	testIntegerObject(t, i, 6)
}

func TestBuiltinStr(t *testing.T) {
	input := `str(1)`
	evaluated := testEval(input)

	i, ok := evaluated.(*object.String)
	if !ok {
		newError("str did not return String, got=%T", evaluated.Type())
	}

	testStringObject(t, i, "1")
}

/*
Operator Testing
*/
func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`
	evaluated := testEval(input)

	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

/*
Value Validation
*/
func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
		return false
	}
	return true
}

func testFloatObject(t *testing.T, obj object.Object, expected float64) bool {
	result, ok := obj.(*object.Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%f, want=%f",
			result.Value, expected)
		return false
	}
	return true
}

func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("object is not String. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("string has wrong value. got=%s, want=%s",
			result.Value, expected)
		return false
	}
	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
		return false
	}

	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

/*
Error Handling
*/
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
if (10 > 1):
	true + false`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
if (10 > 1):
	if (10 > 1):
		return true + false
	return 1`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
		{
			`
def f(x):
	return x
{"name": "Monkey"}[f];`,
			"unusable as hash key: FUNCTION",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)",
				evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMessage, errObj.Message)
		}
	}
}
