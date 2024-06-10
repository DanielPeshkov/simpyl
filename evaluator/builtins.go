package evaluator

import (
	"fmt"
	"math"
	"simpyl/algorithms"
	"simpyl/object"
	"slices"
	"strings"
	"unicode"
)

var builtins = map[string]*object.Builtin{
	"print": {
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return NULL
		},
	},
	"len": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			switch arg := args[0].(type) {
			case *object.List:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s",
					args[0].Type())
			}
		},
	},
	"range": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newError("range function takes integer type, got=%T", args[0].Type())
			}
			r := args[0].(*object.Integer).Value
			elements := make([]object.Object, r)
			for i := range r {
				elements[i] = &object.Integer{Value: i}
			}

			return &object.List{Elements: elements}
		},
	},
	"min": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("min function takes list type, got=%T", args[0].Type())
			}

			vals := args[0].(*object.List).Elements
			if len(vals) == 0 {
				return newError("cannot take min of empty list")
			}

			m := float64(0)
			v := float64(0)
			floatFlag := false

			if vals[0].Type() == object.INTEGER_OBJ {
				m = float64(vals[0].(*object.Integer).Value)
			} else {
				m = vals[0].(*object.Float).Value
				floatFlag = true
			}

			for i := range vals {
				if vals[i].Type() == object.INTEGER_OBJ {
					v = float64(vals[i].(*object.Integer).Value)
				} else if vals[i].Type() == object.FLOAT_OBJ {
					v = vals[i].(*object.Float).Value
					floatFlag = true
				} else {
					return newError("min function requires Integer or Float type, got=%T", vals[i].Type())
				}
				m = min(m, v)
			}

			if floatFlag {
				return &object.Float{Value: m}
			}
			return &object.Integer{Value: int64(m)}
		},
	},
	"max": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("max function takes list type, got=%T", args[0].Type())
			}
			vals := args[0].(*object.List).Elements
			if len(vals) == 0 {
				return newError("cannot take max of empty list")
			}

			m := float64(0)
			v := float64(0)
			floatFlag := false

			if vals[0].Type() == object.INTEGER_OBJ {
				m = float64(vals[0].(*object.Integer).Value)
			} else {
				m = vals[0].(*object.Float).Value
				floatFlag = true
			}

			for i := range vals {
				if vals[i].Type() == object.INTEGER_OBJ {
					v = float64(vals[i].(*object.Integer).Value)
				} else if vals[i].Type() == object.FLOAT_OBJ {
					v = vals[i].(*object.Float).Value
					floatFlag = true
				} else {
					return newError("min function requires Integer or Float type, got=%T", vals[i].Type())
				}
				m = max(m, v)
			}

			if floatFlag {
				return &object.Float{Value: m}
			}
			return &object.Integer{Value: int64(m)}
		},
	},
	"abs": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() == object.INTEGER_OBJ {
				n := args[0].(*object.Integer).Value
				y := n >> 63

				return &object.Integer{Value: (n ^ y) - y}
			}
			if args[0].Type() == object.FLOAT_OBJ {
				n := args[0].(*object.Float).Value
				if n < 0 {
					n = -n
				}

				return &object.Float{Value: n}
			}
			return newError("abs function takes Integer or Float type, got=%T", args[0].Type())
		},
	},
	"sum": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("sum function takes list type, got=%T", args[0].Type())
			}
			vals := args[0].(*object.List).Elements
			if len(vals) == 0 {
				return newError("cannot take sum of empty list")
			}

			n := float64(0)
			v := float64(0)
			floatFlag := false

			for i := range vals {
				if vals[i].Type() == object.INTEGER_OBJ {
					v = float64(vals[i].(*object.Integer).Value)
				} else if vals[i].Type() == object.FLOAT_OBJ {
					v = vals[i].(*object.Float).Value
					floatFlag = true
				} else {
					return newError("max function requires Integer or Float type, got=%T", vals[i].Type())
				}
				n += v
			}

			if floatFlag {
				return &object.Float{Value: n}
			}
			return &object.Integer{Value: int64(n)}
		},
	},
	"str": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("str() takes 1 argument. got=%d, want=1",
					len(args))
			}

			return &object.String{Value: args[0].Inspect()}
		},
	},
	"reversed": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("sum function takes list type, got=%T", args[0].Type())
			}
			vals := args[0].(*object.List).Elements
			if len(vals) == 0 {
				return args[0]
			}
			list := args[0].(*object.List).Elements
			slices.Reverse(list)
			return &object.List{Elements: list}
		},
	},
	"round": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() == object.INTEGER_OBJ {
				return args[0]
			}
			if args[0].Type() == object.FLOAT_OBJ {
				n := math.Round(args[0].(*object.Float).Value)

				return &object.Integer{Value: int64(n)}
			}
			return newError("round function takes Integer or Float type, got=%T", args[0].Type())
		},
	},
	"sorted": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("sorted() takes list type, got=%T", args[0].Type())
			}

			list := algorithms.MergeSort(args[0].(*object.List).Elements)
			return &object.List{Elements: list}
		},
	},
	"list": {
		Fn: func(args ...object.Object) object.Object {
			list := &object.List{}

			if len(args) == 1 {
				if args[0].Type() == object.SET_OBJ {
					set := args[0].(*object.Set)
					for _, item := range set.Values {
						list.Elements = append(list.Elements, item)
					}
				} else {
					list.Elements = append(list.Elements, args...)
				}
			} else {
				list.Elements = append(list.Elements, args...)
			}

			return list
		},
	},
	"dict": {
		Fn: func(args ...object.Object) object.Object {
			dict := &object.Dict{}

			return dict
		},
	},
	"set": {
		Fn: func(args ...object.Object) object.Object {
			set := &object.Set{Values: make(map[object.HashKey]object.Object)}

			for _, arg := range args {
				hashKey, ok := arg.(object.Hashable)
				if !ok {
					return newError("argument cannot be hashed: %s", arg.Type())
				}
				key := hashKey.HashKey()

				if _, ok := set.Values[key]; !ok {
					set.Values[key] = arg
				}
			}

			return set
		},
	},
}

var listMethods = map[string]*object.BuiltinMethod{
	"append": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}

			if obj.Type() != object.LIST_OBJ {
				return newError("list.append() must be called on list, got %s",
					args[0].Type())
			}

			list := obj.(*object.List)
			list.Elements = append(list.Elements, args[0])

			return list
		},
	},
	"reverse": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("list.reverse() takes no arguments")
			}

			if obj.Type() != object.LIST_OBJ {
				return newError("list.reverse() must be called on list, got %s",
					args[0].Type())
			}

			list := obj.(*object.List)
			slices.Reverse(list.Elements)

			return NULL
		},
	},
	"copy": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("list.copy() takes no arguments")
			}

			if obj.Type() != object.LIST_OBJ {
				return newError("list.copy() must be called on list, got %s",
					args[0].Type())
			}

			src := obj.(*object.List)
			list := make([]object.Object, len(src.Elements))
			copy(list, src.Elements)

			return &object.List{Elements: list}
		},
	},
	"pop": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			ind := -1
			if len(args) == 1 {
				arg, ok := args[0].(*object.Integer)
				if !ok {
					return newError("index of list.pop() must be Integer type, got=%T", args[0])
				}
				ind = int(arg.Value)
			} else if len(args) > 1 {
				return newError("list.pop() take at most 1 index, got=%d",
					len(args))
			}

			if obj.Type() != object.LIST_OBJ {
				return newError("list.pop() must be called on list, got %s",
					args[0].Type())
			}

			list := obj.(*object.List)
			elements := list.Elements
			if ind < 0 {
				ind = len(elements) + ind
			}
			if ind >= len(elements) {
				return newError("index out of range of list.pop()")
			}

			result := elements[ind]
			newList := make([]object.Object, len(elements)-1)
			if len(newList) != len(elements)-1 {
				return newError("incorrect size for return list")
			}
			if ind == len(elements) {
				newList = elements[:ind]
			} else {
				newList = append(elements[:ind], elements[ind+1:]...)
			}
			list.Elements = newList

			return result
		},
	},
	"sort": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("list.sort() takes no arguments")
			}

			if obj.Type() != object.LIST_OBJ {
				return newError("list.sort() must be called on list, got %s",
					args[0].Type())
			}

			list := obj.(*object.List)
			list.Elements = algorithms.MergeSort(list.Elements)

			return NULL
		},
	},
}

var stringMethods = map[string]*object.BuiltinMethod{
	"join": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}

			if obj.Type() != object.STRING_OBJ {
				return newError("string.join() must be called on string, got %s",
					obj.Type())
			}

			if args[0].Type() != object.LIST_OBJ {
				return newError("string.join() takes a list, got %s",
					args[0].Type())
			}

			str := obj.(*object.String).Value
			list := args[0].(*object.List).Elements

			if len(list) == 0 {
				return newError("cannot join empty list")
			}

			result := list[0].Inspect()
			if len(list) == 1 {
				return &object.String{Value: result}
			}

			for _, el := range list[1:] {
				result += str
				result += el.Inspect()
			}

			return &object.String{Value: result}
		},
	},
	"upper": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("string.upper() takes no arguments")
			}

			if obj.Type() != object.STRING_OBJ {
				return newError("string.upper() must be called on string, got %s",
					obj.Type())
			}

			str := obj.(*object.String).Value
			str = strings.ToUpper(str)

			return &object.String{Value: str}
		},
	},
	"lower": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("string.lower() takes no arguments")
			}

			if obj.Type() != object.STRING_OBJ {
				return newError("string.lower() must be called on string, got %s",
					obj.Type())
			}

			str := obj.(*object.String).Value
			str = strings.ToLower(str)

			return &object.String{Value: str}
		},
	},
	"isupper": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("string.isupper() takes no arguments")
			}

			if obj.Type() != object.STRING_OBJ {
				return newError("string.isupper() must be called on string, got %s",
					obj.Type())
			}

			str := obj.(*object.String).Value
			for _, r := range str {
				if !unicode.IsUpper(r) && unicode.IsLetter(r) {
					return FALSE
				}
			}

			return TRUE
		},
	},
	"islower": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("string.islower() takes no arguments")
			}

			if obj.Type() != object.STRING_OBJ {
				return newError("string.islower() must be called on string, got %s",
					obj.Type())
			}

			str := obj.(*object.String).Value
			for _, r := range str {
				if !unicode.IsLower(r) && unicode.IsLetter(r) {
					return FALSE
				}
			}

			return TRUE
		},
	},
	"swapcase": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("string.swapcase() takes no arguments")
			}

			if obj.Type() != object.STRING_OBJ {
				return newError("string.swapcase() must be called on string, got %s",
					obj.Type())
			}

			str := obj.(*object.String).Value
			str = strings.Map(func(r rune) rune {
				switch {
				case unicode.IsLower(r):
					return unicode.ToUpper(r)
				case unicode.IsUpper(r):
					return unicode.ToLower(r)
				}
				return r
			}, str)

			return &object.String{Value: str}
		},
	},
}

var dictMethods = map[string]*object.BuiltinMethod{
	"keys": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("dict.keys() takes no arguments")
			}

			if obj.Type() != object.DICT_OBJ {
				return newError("dict.keys() must be called on dict, got %s",
					args[0].Type())
			}

			keys := &object.List{}

			dict := obj.(*object.Dict)
			for _, p := range dict.Pairs {
				keys.Elements = append(keys.Elements, p.Key)
			}

			return keys
		},
	},
	"values": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("dict.values() takes no arguments")
			}

			if obj.Type() != object.DICT_OBJ {
				return newError("dict.values() must be called on dict, got %s",
					args[0].Type())
			}

			values := &object.List{}

			dict := obj.(*object.Dict)
			for _, p := range dict.Pairs {
				values.Elements = append(values.Elements, p.Value)
			}

			return values
		},
	},
	"items": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("dict.items() takes no arguments")
			}

			if obj.Type() != object.DICT_OBJ {
				return newError("dict.items() must be called on dict, got %s",
					args[0].Type())
			}

			items := &object.List{}

			dict := obj.(*object.Dict)
			for _, p := range dict.Pairs {
				elements := make([]object.Object, 2)
				elements[0] = p.Key
				elements[1] = p.Value
				item := &object.List{Elements: elements}
				items.Elements = append(items.Elements, item)
			}

			return items
		},
	},
	"pop": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("dict.pop() requires key as argument")
			}

			if obj.Type() != object.DICT_OBJ {
				return newError("dict.pop() must be called on list, got %s",
					args[0].Type())
			}

			dict := obj.(*object.Dict)

			hashKey, ok := args[0].(object.Hashable)
			if !ok {
				return newError("unusable as hash key: %s", args[0].Type())
			}

			result, ok := dict.Pairs[hashKey.HashKey()]
			if !ok {
				return newError("%s not found in dict", args[0].Inspect())
			}

			delete(dict.Pairs, hashKey.HashKey())

			return result.Value
		},
	},
}

var setMethods = map[string]*object.BuiltinMethod{
	"add": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("set.add() takes 1 argument, got=%d", len(args))
			}

			if obj.Type() != object.SET_OBJ {
				return newError("set.add() must be called on set, got %s",
					args[0].Type())
			}
			set := obj.(*object.Set)

			hashKey, ok := args[0].(object.Hashable)
			if !ok {
				return newError("argument cannot be hashed: %s", args[0].Type())
			}
			key := hashKey.HashKey()

			if _, ok := set.Values[key]; ok {
				return NULL
			}

			set.Values[key] = args[0]

			return NULL
		},
	},
	"remove": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("set.remove() takes 1 argument, got=%d", len(args))
			}

			if obj.Type() != object.SET_OBJ {
				return newError("set.remove() must be called on set, got %s",
					args[0].Type())
			}
			set := obj.(*object.Set)

			hashKey, ok := args[0].(object.Hashable)
			if !ok {
				return newError("argument cannot be hashed: %s", args[0].Type())
			}
			key := hashKey.HashKey()

			_, ok = set.Values[key]
			if !ok {
				return newError("%s not found in set", args[0].Inspect())
			}

			delete(set.Values, key)

			return NULL
		},
	},
	"discard": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("set.discard() takes 1 argument, got=%d", len(args))
			}

			if obj.Type() != object.SET_OBJ {
				return newError("set.discard() must be called on set, got %s",
					args[0].Type())
			}
			set := obj.(*object.Set)

			hashKey, ok := args[0].(object.Hashable)
			if !ok {
				return newError("argument cannot be hashed: %s", args[0].Type())
			}
			key := hashKey.HashKey()

			delete(set.Values, key)

			return NULL
		},
	},
	"pop": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("set.pop() takes 1 argument, got=%d", len(args))
			}

			if obj.Type() != object.SET_OBJ {
				return newError("set.pop() must be called on set, got %s",
					args[0].Type())
			}
			set := obj.(*object.Set)

			hashKey, ok := args[0].(object.Hashable)
			if !ok {
				return newError("argument cannot be hashed: %s", args[0].Type())
			}
			key := hashKey.HashKey()

			result, ok := set.Values[key]
			if !ok {
				return newError("%s not found in set", args[0].Inspect())
			}

			delete(set.Values, key)

			return result
		},
	},
	"intersection": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("set.intersection() takes 1 argument, got=%d", len(args))
			}
			if args[0].Type() != object.SET_OBJ {
				return newError("set.intersection() takes set as argument, got %s",
					args[0].Type())
			}

			if obj.Type() != object.SET_OBJ {
				return newError("set.intersection() must be called on set, got %s",
					obj.Type())
			}
			a := obj.(*object.Set)
			b := args[0].(*object.Set)
			result := &object.Set{Values: make(map[object.HashKey]object.Object)}

			for hash, val := range a.Values {
				cpy, ok := b.Values[hash]
				if ok {
					if cpy.Inspect() == val.Inspect() {
						if _, ok := result.Values[hash]; !ok {
							result.Values[hash] = val
						}
					}
				}
			}
			for hash, val := range b.Values {
				cpy, ok := a.Values[hash]
				if ok {
					if cpy.Inspect() == val.Inspect() {
						if _, ok := result.Values[hash]; !ok {
							result.Values[hash] = val
						}
					}
				}
			}

			return result
		},
	},
	"union": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("set.union() takes 1 argument, got=%d", len(args))
			}
			if args[0].Type() != object.SET_OBJ {
				return newError("set.union() takes set as argument, got %s",
					args[0].Type())
			}

			if obj.Type() != object.SET_OBJ {
				return newError("set.union() must be called on set, got %s",
					obj.Type())
			}
			a := obj.(*object.Set)
			b := args[0].(*object.Set)
			result := &object.Set{Values: make(map[object.HashKey]object.Object)}

			for hash, val := range a.Values {
				if _, ok := result.Values[hash]; !ok {
					result.Values[hash] = val
				}
			}
			for hash, val := range b.Values {
				if _, ok := result.Values[hash]; !ok {
					result.Values[hash] = val
				}
			}

			return result
		},
	},
	"difference": {
		Fn: func(obj object.Object, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("set.difference() takes 1 argument, got=%d", len(args))
			}
			if args[0].Type() != object.SET_OBJ {
				return newError("set.difference() takes set as argument, got %s",
					args[0].Type())
			}

			if obj.Type() != object.SET_OBJ {
				return newError("set.difference() must be called on set, got %s",
					obj.Type())
			}
			a := obj.(*object.Set)
			b := args[0].(*object.Set)
			result := &object.Set{Values: make(map[object.HashKey]object.Object)}

			for hash, val := range a.Values {
				cpy, ok := b.Values[hash]
				if ok {
					if cpy.Inspect() != val.Inspect() {
						if _, ok := result.Values[hash]; !ok {
							result.Values[hash] = val
						}
					}
				} else {
					if _, ok := result.Values[hash]; !ok {
						result.Values[hash] = val
					}
				}
			}
			for hash, val := range b.Values {
				cpy, ok := a.Values[hash]
				if ok {
					if cpy.Inspect() != val.Inspect() {
						if _, ok := result.Values[hash]; !ok {
							result.Values[hash] = val
						}
					}
				} else {
					if _, ok := result.Values[hash]; !ok {
						result.Values[hash] = val
					}
				}
			}

			return result
		},
	},
}
