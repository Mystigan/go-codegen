package main

import (
	"flag"
	"fmt"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	. "github.com/dave/jennifer/jen"
	"golang.org/x/tools/go/packages"
)

func main() {
	structName := flag.String("type", "", "the name of the struct to target")
	pkgName := flag.String("pkg", os.Getenv("GOPACKAGE"), "the target package name, defaults to the file with go:generate comment")
	out := flag.String("out", os.Getenv("GOFILE"), "the output directory")
	flag.Parse()

	path, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get package directory: %v", err)
	}
	pkg := loadPackage(path)
	obj := pkg.Types.Scope().Lookup(*structName)
	if obj == nil {
		log.Fatalf("struct %s not found", *structName)
	}

	// Check if it is a declared type.
	if _, ok := obj.(*types.TypeName); !ok {
		log.Fatalf("%v is not a named type", obj)
	}

	// Check if the type is a struct.
	structType, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		log.Fatalf("%v is not a struct", obj)
	}

	// Generate code.
	if err := generate(*structName, structType, *pkgName, *out); err != nil {
		panic(err)
	}
}

func generate(structName string, structType *types.Struct, goPackage, goFile string) error {
	// Start a new file in this package.
	f := NewFile(goPackage)

	// Add a comment.
	f.PackageComment("Code generated by generator, DO NOT EDIT.")
	type sortableCode struct {
		vars      *types.Var
		code      Code
		fieldName string
		fieldType string
	}
	var fields []sortableCode
	var lets []Code
	letsMapper := make(map[string]bool)

	// Build the struct fields.
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		//tagValue := structType.Tag(i)
		// Make name private
		fieldName := lowerFirst(field.Name())

		// TODO: Lowercase words like ID, JSON.
		if strings.ToUpper(field.Name()) == field.Name() {
			fieldName = strings.ToLower(field.Name())
		}
		code := Id(fieldName)

		var fieldType string
		switch v := field.Type().(type) {
		case *types.Basic:
			code.Id(v.String())
			fieldType = v.String()
		case *types.Pointer:
			code.Id(v.String())
			fieldType = v.String()
		case *types.Named:
			typeName := v.Obj()
			switch typeName.Pkg().Name() {
			case "sql":
				// NullTime, NullString ...
				switch name := typeName.Name(); name {
				case "NullTime":
					code.Qual("time", "Time")
					fieldType = "time.Time"

					// Handle initialization of sql NullTypes
					// var confirmedAt time.Time
					// if u.ConfirmedAt.Valid {
					//  confirmedAt = u.confirmedAt.Time
					// }
					let := Var().Id(fieldName).Qual("time", "Time").Line()
					let.If().Id("u").Dot(field.Name()).Dot("Valid").Block(
						Id(fieldName).Op("=").Id("u").Dot(field.Name()).Dot(strings.Replace(name, "Null", "", -1)),
					)
					lets = append(lets, let)
					letsMapper[fieldName] = true
				default:
					underlyingType := strings.Replace(name, "Null", "", -1)
					code.Id(strings.ToLower(underlyingType))
					fieldType = strings.ToLower(underlyingType)

					let := Var().Id(fieldName).Id(strings.ToLower(underlyingType)).Line()
					let.If().Id("u").Dot(field.Name()).Dot("Valid").Block(
						Id(fieldName).Op("=").Id("u").Dot(field.Name()).Dot(underlyingType),
					)
					lets = append(lets, let)
					letsMapper[fieldName] = true
				}
			default:
				// Qual import packages.
				code.Qual(
					typeName.Pkg().Path(),
					typeName.Name(),
				)
				fieldType = strings.Join([]string{
					typeName.Pkg().Name(),
					typeName.Name(),
				}, ".")
			}
		case *types.Slice:
			code.Id(v.String())
			fieldType = v.String()
		default:
			return fmt.Errorf("struct field type not handled: %T", v)
		}
		fields = append(fields, sortableCode{
			vars:      field,
			code:      code,
			fieldName: fieldName,
			fieldType: fieldType,
		})
	}

	var sortable bool
	sortable = true
	if sortable {
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].fieldName < fields[j].fieldName
		})
	}

	codes := make([]Code, len(fields))
	for i := range fields {
		codes[i] = fields[i].code
	}

	// Generate a struct with the fields.
	f.Type().Id(structName).Struct(codes...)

	dict := Dict{}
	for i := range fields {
		name := fields[i].fieldName
		dict[Id(name)] = Id(name)
	}

	mapperDict := Dict{}
	for i := range fields {
		field := fields[i]
		name := field.fieldName
		prev := field.vars.Name()
		if letsMapper[name] {
			mapperDict[Id(name)] = Id(name)
		} else {
			mapperDict[Id(name)] = Id("u").Dot(prev)
		}
	}

	// Generate the constructor for that field.
	f.Func().
		Id("New" + structName). // Function name.
		Params(codes...).       // Args.
		Id(structName).         // Return type.
		Block(
			Return(Id(structName).Values(dict)),
		).Line()

	// Generate the adapter for that field.
	f.Func().
		Id("NewFrom" + structName).
		Params(Id("u").
			Qual(os.Getenv("GOPACKAGE"), structName)).
		Id(structName).Block(
		append(lets, Return(Id(structName).Values(mapperDict)))...,
	).Line()

	// Generate getter methods.
	for i := range fields {
		field := fields[i]
		prefix := strings.ToLower(string(structName[0]))
		f.Func().Params(Id(prefix).Id(structName)).Id(field.vars.Name()).
			Params().
			Id(field.fieldType).Block(
			Return(Id(prefix).Dot(field.fieldName)),
		).Line()
	}

	ext := filepath.Ext(goFile)
	baseFilename := goFile[0 : len(goFile)-len(ext)]
	targetFilename := baseFilename + "_gen.go"
	return f.Save(targetFilename)
}

func loadPackage(path string) *packages.Package {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		log.Fatalf("failed to load package: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}
	return pkgs[0]
}

// lowerFirst converts the first character to lowercase.
func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
