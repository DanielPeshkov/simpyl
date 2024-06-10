package evaluator

import (
	"fmt"
	"simpyl/ast"
	"simpyl/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// Statements
	case *ast.Program:
		return evalProgram(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)

	case *ast.FunctionStatement:
		params := node.Parameters
		body := node.Body
		name := node.Name

		env.Set(name, &object.Function{Parameters: params, Env: env, Body: body, Name: name})

	case *ast.ForStatement:
		evalForLoop(node, env)

	case *ast.WhileStatement:
		evalWhileLoop(node, env)

	// Expressions
	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.ListLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.List{Elements: elements}

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		colon := node.Colon
		if colon {
			end := Eval(node.EndIndex, env)
			return evalIndexExpression(left, index, colon, end)
		}
		return evalIndexExpression(left, index, colon, nil)

	case *ast.IndexAssignExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		val := Eval(node.Value, env)
		return evalIndexAssignExpression(left, index, val)

	case *ast.DictLiteral:
		return evalDictLiteral(node, env)

	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)

	case *ast.ObjectMethod:
		obj := Eval(node.Obj, env)
		if isError(obj) {
			return obj
		}

		method, ok := node.Method.(*ast.CallExpression)
		if !ok {
			newError("Object method not ast.CallExpression. got=%T", node.Method)
		}

		args := evalExpressions(method.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyObjectMethod(obj, method.Function, args)

	case *ast.InExpression:
		left := Eval(node.Left, env)
		right := Eval(node.Right, env)
		return evalInExpression(left, right)

	// Operators
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		right := Eval(node.Right, env)
		if isError(left) {
			return left
		}
		if isError(right) {
			return right
		}

		return evalInfixExpression(node.Operator, left, right)
	}
	return nil
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {

	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		return fn.Fn(args...)

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalIndexExpression(left, index object.Object, colon bool, end object.Object) object.Object {
	switch {
	case left.Type() == object.LIST_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalListIndexExpression(left, index, colon, end)
	case left.Type() == object.DICT_OBJ:
		return evalDictIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalIndexAssignExpression(left, index object.Object, val object.Object) object.Object {
	switch {
	case left.Type() == object.LIST_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalListIndexAssignExpression(left, index, val)
	case left.Type() == object.DICT_OBJ:
		return evalDictIndexAssignExpression(left, index, val)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalListIndexExpression(list, index object.Object, colon bool, end object.Object) object.Object {
	listObject := list.(*object.List)
	idx := index.(*object.Integer).Value

	// Allows negative indexing from end of list
	if idx < 0 {
		idx = int64(len(listObject.Elements)) + idx
	}

	if colon {
		edx := end.(*object.Integer).Value
		if edx < 0 {
			edx = int64(len(listObject.Elements)) + edx
		}

		if edx <= idx {
			return newError("Starting index must be before ending index")
		}

		returnList := &object.List{}
		returnList.Elements = listObject.Elements[idx:edx]

		return returnList
	}

	max := int64(len(listObject.Elements) - 1)
	if idx < 0 || idx > max {
		return NULL
	}
	return listObject.Elements[idx]
}

func evalListIndexAssignExpression(list, index, val object.Object) object.Object {
	listObject := list.(*object.List)
	idx := index.(*object.Integer).Value

	// Allows negative indexing from end of list
	if idx < 0 {
		idx = int64(len(listObject.Elements)) + idx
	}

	max := int64(len(listObject.Elements) - 1)
	if idx < 0 || idx > max {
		return NULL
	}

	listObject.Elements[idx] = val

	return listObject
}

func evalDictLiteral(node *ast.DictLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Dict{Pairs: pairs}
}

func evalDictIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Dict)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalDictIndexAssignExpression(dict, index, val object.Object) object.Object {
	dictObject := dict.(*object.Dict)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as dict key: %s", index.Type())
	}

	dictObject.Pairs[key.HashKey()] = object.HashPair{Key: index, Value: val}

	return dictObject
}

func applyObjectMethod(obj object.Object, method ast.Expression, args []object.Object) object.Object {
	switch obj.(type) {

	case *object.List:
		return listMethods[method.String()].Fn(obj, args...)

	case *object.String:
		return stringMethods[method.String()].Fn(obj, args...)

	case *object.Dict:
		return dictMethods[method.String()].Fn(obj, args...)

	case *object.Set:
		return setMethods[method.String()].Fn(obj, args...)

	default:
		return newError("not a function: %s", obj.Type())
	}
}

func evalInExpression(left, right object.Object) object.Object {
	switch right.Type() {
	case "LIST":
		return searchList(left, right)

	case "SET":
		return searchSet(left, right)

	default:
		return newError("cannot check if object of type %T contains an object", right.Type())
	}
}

func searchSet(target, obj object.Object) object.Object {
	set := obj.(*object.Set)

	hashKey, ok := target.(object.Hashable)
	if !ok {
		return newError("object cannot be hashed: %s", target.Type())
	}
	key := hashKey.HashKey()

	_, ok = set.Values[key]
	if ok {
		return TRUE
	}

	return FALSE
}

func searchList(target, obj object.Object) object.Object {
	list := obj.(*object.List)
	valType := target.Type()

	for _, el := range list.Elements {
		if el.Type() == valType {
			switch target.(type) {
			case *object.Integer:
				if target.(*object.Integer).Value == el.(*object.Integer).Value {
					return TRUE
				}

			case *object.Float:
				if target.(*object.Float).Value == el.(*object.Float).Value {
					return TRUE
				}

			case *object.Boolean:
				if target.(*object.Boolean).Value == el.(*object.Boolean).Value {
					return TRUE
				}

			case *object.String:
				if target.(*object.String).Value == el.(*object.String).Value {
					return TRUE
				}

			case *object.List:
				if compareListsEqual(target.(*object.List), el.(*object.List)).Value {
					return TRUE
				}

			case *object.Set:
				if compareSetsEqual(target.(*object.Set), el.(*object.Set)).Value {
					return TRUE
				}

			default:
				return newError("cannot search list for object of type %T", target)
			}
		}
	}

	return FALSE
}

func compareListsEqual(l1, l2 *object.List) object.Boolean {
	if len(l1.Elements) != len(l2.Elements) {
		return *FALSE
	}

	for i := range l1.Elements {
		if l1.Elements[i].Inspect() != l2.Elements[i].Inspect() {
			return *FALSE
		}
	}

	return *TRUE
}

func compareSetsEqual(s1, s2 *object.Set) object.Boolean {
	for key, val := range s1.Values {
		if s2.Values[key].Inspect() != val.Inspect() {
			return *FALSE
		}
	}
	for key, val := range s2.Values {
		if s1.Values[key].Inspect() != val.Inspect() {
			return *FALSE
		}
	}

	return *TRUE
}

/*
Loop Statements
*/
func evalForLoop(node *ast.ForStatement, env *object.Environment) {
	exp := Eval(node.Iterable, env)
	iterable, ok := exp.(*object.List)
	if !ok {
		newError("Iterable passed to for loop must be list, got=%T", exp)
	}

	block := node.Body

	iterator := node.Iterator.Value
	for _, i := range iterable.Elements {
		env.Set(iterator, i)
		evalBlockStatement(block, env)
	}
}

func evalWhileLoop(node *ast.WhileStatement, env *object.Environment) {
	exp := Eval(node.Condition, env)

	for isTruthy(exp) {
		evalBlockStatement(node.Body, env)
		exp = Eval(node.Condition, env)
	}
}

/*
Conditional Expressions
*/
func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)

	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func isTruthy(obj object.Object) bool {
	switch {
	case obj == NULL:
		return false
	case obj == TRUE:
		return true
	case obj == FALSE:
		return false
	case obj.Type() == object.INTEGER_OBJ:
		obj := obj.(*object.Integer)
		if obj.Value == 0 {
			return false
		} else {
			return true
		}

	default:
		return true
	}
}

/*
Prefix Expressions
*/
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch right.Type() {
	case object.INTEGER_OBJ:
		value := right.(*object.Integer).Value
		return &object.Integer{Value: -value}

	case object.FLOAT_OBJ:
		value := right.(*object.Float).Value
		return &object.Float{Value: -value}

	default:
		return newError("unknown operator: -%s", right.Type())
	}
}

/*
Infix Expressions
*/
func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ || right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())

	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {

	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())

	}
}

func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {

	leftVal := float64(0)
	rightVal := float64(0)

	if left.Type() == object.FLOAT_OBJ {
		leftVal = left.(*object.Float).Value
	} else if left.Type() == object.INTEGER_OBJ {
		leftVal = float64(left.(*object.Integer).Value)
	}

	if right.Type() == object.FLOAT_OBJ {
		rightVal = right.(*object.Float).Value
	} else if right.Type() == object.INTEGER_OBJ {
		rightVal = float64(right.(*object.Integer).Value)
	}

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())

	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {

	switch operator {
	case "+":
		leftVal := left.(*object.String).Value
		rightVal := right.(*object.String).Value
		return &object.String{Value: leftVal + rightVal}

	case "==":
		leftVal := left.(*object.String).Value
		rightVal := right.(*object.String).Value
		val := leftVal == rightVal
		return &object.Boolean{Value: val}

	case "!=":
		leftVal := left.(*object.String).Value
		rightVal := right.(*object.String).Value
		val := leftVal != rightVal
		return &object.Boolean{Value: val}

	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}

}

/*
Error Handling
*/
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
