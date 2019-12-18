package fixturegenerator

import (
	"fmt"
	"go/ast"
	"math/rand"
	"os"
	"path/filepath"
)

const (
	maxInt        int = 1000000
	maxListLength int = 3
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// GetStructFixture returns a fixture for a struct. It's the entrypoint to the fixture generator.
func GetStructFixture(structType *structType, imports map[string]string, pointer bool, indentation int) string {
	var fixture string
	if pointer {
		fixture = fmt.Sprintf("&%s{\n", structType.name)
	} else {
		fixture = fmt.Sprintf("%s{\n", structType.name)
	}

	for _, field := range structType.fields {
		value, err := getValueFixture(field.goType, imports, pointer, indentation+1)
		if err != nil {
			// Best effort - even if some fields can not be generated it still saves some time.
			fmt.Println(err)
			continue
		}

		fixture += fmt.Sprintf("%s%s: %v,\n", getTabs(indentation), field.name, value)
	}

	fixture += fmt.Sprintf("%s}", getTabs(indentation-1))

	return fixture
}

func getValueFixture(goType ast.Expr, imports map[string]string, pointer bool, indentation int) (string, error) {
	switch concreteType := goType.(type) {
	case *ast.Ident:
		if concreteType.Obj != nil {
			// We have a struct here.
			return getStructFieldFixture(concreteType.Obj, imports, pointer, indentation)
		}
		return getSimpleTypeFixture(concreteType.Name)
	case *ast.ArrayType:
		return getArrayFixture(concreteType, imports, pointer, indentation)
	case *ast.MapType:
		return getMapFixture(concreteType, imports, pointer, indentation)
	case *ast.SelectorExpr:
		// External pkg type.
		return getExternalStructFieldFixture(concreteType, imports, pointer, indentation)
	case *ast.StarExpr:
		// Pointer.
		return getValueFixture(concreteType.X, imports, true, indentation)
	}

	return "", fmt.Errorf("Can not generate field data")
}

func getExternalStructFieldFixture(selectorExpr *ast.SelectorExpr, imports map[string]string, pointer bool, indentation int) (string, error) {
	expr, _ := selectorExpr.X.(*ast.Ident)
	packagePath := getPackagePath(imports[expr.Name])
	structType, err := GetStruct(packagePath, selectorExpr.Sel.Name)
	if err != nil {
		return "", err
	}
	newImports, err := GetImports(packagePath)
	if err != nil {
		return "", err
	}
	return GetStructFixture(structType, newImports, pointer, indentation), nil
}

func getStructFieldFixture(object *ast.Object, imports map[string]string, pointer bool, indentation int) (string, error) {
	spec, ok := object.Decl.(*ast.TypeSpec)
	if !ok {
		return "", fmt.Errorf("Not a type")
	}

	structType, err := typeSpecToStructType(spec)
	if err != nil {
		return "", err
	}

	return GetStructFixture(structType, imports, pointer, indentation), nil
}

func getArrayFixture(arrayType *ast.ArrayType, imports map[string]string, pointer bool, indentation int) (string, error) {
	count := int(rand.Uint32())
	if count == 0 {
		// We should have at least one element.
		count = 1
	}
	if count > maxListLength {
		// Control number of elements.
		count = maxListLength
	}
	typeName, err := getTypeName(arrayType)
	if err != nil {
		return "", err
	}
	fixture := fmt.Sprintf("%s{", typeName)
	for i := 0; i < count-1; i++ {
		value, err := getValueFixture(arrayType.Elt, imports, pointer, indentation)
		if err != nil {
			return "", err
		}
		fixture += value
		fixture += ", "
	}
	value, err := getValueFixture(arrayType.Elt, imports, pointer, indentation)
	if err != nil {
		return "", err
	}
	fixture += value
	fixture += "}"

	return fixture, nil
}

func getMapFixture(mapType *ast.MapType, imports map[string]string, pointer bool, indentation int) (string, error) {
	count := int(rand.Uint32())
	if count == 0 {
		// We should have at least one element.
		count = 1
	}
	if count > maxListLength {
		// Control number of elements.
		count = maxListLength
	}
	typeName, err := getTypeName(mapType)
	if err != nil {
		return "", err
	}

	fixture := fmt.Sprintf("%s{\n", typeName)
	for i := 0; i < count-1; i++ {
		key, err := getValueFixture(mapType.Key, imports, pointer, indentation+1)
		if err != nil {
			return "", err
		}

		value, err := getValueFixture(mapType.Value, imports, pointer, indentation+1)
		if err != nil {
			return "", err
		}
		fixture += fmt.Sprintf("%s%s: %s,\n", getTabs(indentation), key, value)
	}

	key, err := getValueFixture(mapType.Key, imports, pointer, indentation+1)
	if err != nil {
		return "", err
	}

	value, err := getValueFixture(mapType.Value, imports, pointer, indentation+1)
	if err != nil {
		return "", err
	}
	fixture += fmt.Sprintf("%s%s: %s}", getTabs(indentation), key, value)

	return fixture, nil
}

func getSimpleTypeFixture(concreteType string) (string, error) {
	switch concreteType {
	case "int":
		return getRandomInt(), nil
	case "int8":
		return getRandomInt8(), nil
	case "int16":
		return getRandomInt16(), nil
	case "int32":
		return getRandomInt32(), nil
	case "int64":
		return getRandomInt64(), nil
	case "uint":
		return getRandomUint(), nil
	case "uint8":
		return getRandomUint8(), nil
	case "uint16":
		return getRandomUint16(), nil
	case "uint32":
		return getRandomUint32(), nil
	case "uint64":
		return getRandomUint64(), nil
	case "float32":
		return getRandomFloat32(), nil
	case "float64":
		return getRandomFloat64(), nil
	case "string":
		return getRandomString(), nil
	case "bool":
		return getRandomBool(), nil
	}

	return "", fmt.Errorf("Unsupported type: %s", concreteType)
}

func getTypeName(goType ast.Expr) (string, error) {
	switch concreteType := goType.(type) {
	case *ast.Ident:
		return concreteType.Name, nil
	case *ast.ArrayType:
		elementType, err := getTypeName(concreteType.Elt)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("[]%s", elementType), nil
	case *ast.MapType:
		keyType, err := getTypeName(concreteType.Key)
		if err != nil {
			return "", err
		}
		valueType, err := getTypeName(concreteType.Value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("map[%s]%s", keyType, valueType), nil
	case *ast.StarExpr:
		elementType, err := getTypeName(concreteType.X)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("*%s", elementType), nil
	}

	return "", fmt.Errorf("Can not generate type name")
}


func getPackagePath(path string) string {
	// TODO(Mihai): Add support for multiple paths and for vendoring.
	goPath := os.Getenv("GOPATH")
	wildPath := fmt.Sprintf(`%s/pkg/mod/%s*`, goPath, path)
	s, _ := filepath.Glob(wildPath)
	if len(s) != 0 {
		return s[len(s)-1] // latest version for now
	}
	goRoot := os.Getenv("GOROOT")
	wildPath = fmt.Sprintf(`%s/src/%s*`, goRoot, path)
	s, _ = filepath.Glob(wildPath)
	if len(s) != 0 {
		return s[len(s)-1]
	}
	return fmt.Sprintf(`%s/src/%s*`, goPath, path)
}

func getTabs(indentation int) string {
	tabs := ""
	for i := 0; i < indentation; i++ {
		tabs += "\t"
	}

	return tabs
}

func getRandomInt() string {
	return fmt.Sprintf("%d", rand.Intn(maxInt))
}

func getRandomInt8() string {
	return fmt.Sprintf("int8(%d)", int8(rand.Intn(maxInt)))
}

func getRandomInt16() string {
	return fmt.Sprintf("int16(%d)", int16(rand.Intn(maxInt)))
}

func getRandomInt32() string {
	return fmt.Sprintf("int32(%d)", rand.Int31n(int32(maxInt)))
}

func getRandomInt64() string {
	return fmt.Sprintf("int64(%d)", rand.Int63n(int64(maxInt)))
}

func getRandomUint() string {
	return fmt.Sprintf("uint(%d)", rand.Uint32())
}

func getRandomUint8() string {
	return fmt.Sprintf("uint8(%d)", uint8(rand.Uint32()))
}

func getRandomUint16() string {
	return fmt.Sprintf("uint16(%d)", uint16(rand.Uint32()))
}

func getRandomUint32() string {
	return fmt.Sprintf("uint32(%d)", rand.Uint32())
}

func getRandomUint64() string {
	return fmt.Sprintf("uint64(%d)", rand.Uint64())
}

func getRandomFloat32() string {
	return fmt.Sprintf("float32(%f)", rand.Float32())
}

func getRandomFloat64() string {
	return fmt.Sprintf("float64(%f)", rand.Float64())
}

func getRandomString() string {
	b := make([]rune, rand.Intn(20))
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return fmt.Sprintf("\"%s\"", string(b))
}

func getRandomBool() string {
	return fmt.Sprintf("%v", rand.Intn(10) < 5)
}
