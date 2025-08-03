package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"golang.org/x/tools/go/packages"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <package-path>")
		return
	}

	// 载入包
	pkgPath := os.Args[1]
	processPackage(pkgPath)
}

func processPackage(pkgPath string) {
	cfg := &packages.Config{Mode: packages.LoadAllSyntax}
	pkgs, err := packages.Load(cfg, pkgPath+"/...")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to load packages: %v\n", err)
		return
	}

	// 遍历所有文件
	for _, pkg := range pkgs {
		pkgStructs := []DBStruct{}
		for _, file := range pkg.Syntax {
			fileName := pkg.Fset.File(file.Pos()).Name()
			if strings.HasSuffix(fileName, "_model.go") {
				structs := processFile(pkg, file)
				pkgStructs = append(pkgStructs, structs...)
			}
		}
		if len(pkgStructs) == 0 {
			continue
		}

		dirAbsolutPath := filepath.Dir(pkg.GoFiles[0])
		err := generateCode(dirAbsolutPath, pkg.Name, pkgStructs)
		if err != nil {
			fmt.Printf("failed to generate code: %v\n", err)
		}
	}
}

func processFile(pkg *packages.Package, file *ast.File) []DBStruct {
	return extractStructs(pkg, file)
}

func extractStructs(pkg *packages.Package, file *ast.File) []DBStruct {
	var structs []DBStruct
	ast.Inspect(file, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok {
			structType := pkg.TypesInfo.TypeOf(ts.Type)
			if st, ok := structType.(*types.Struct); ok {
				dbFields := extractDBField(st)
				dbStruct := DBStruct{
					DBName:   ts.Name.Name,
					DBFields: dbFields,
				}
				structs = append(structs, dbStruct)
			}
		}
		return true
	})
	return structs
}

type DBStruct struct {
	DBName   string
	DBFields []DBField
}

type DBField struct {
	Name   string
	DBName string
}

func processAnonymousOrEmbeddedField(f *types.Var) []DBField {
	// This is an embedded field, so we need to get its type
	embeddedStruct, ok := f.Type().Underlying().(*types.Struct)
	if !ok {
		fmt.Printf("%s is not a struct\n", f.Name())
		return []DBField{}
	}
	return extractDBField(embeddedStruct)
}

// extractDBField 从结构体类型中提取 DBField, 仅处理直接定义在结构体中的字段
func extractDBField(structType *types.Struct) []DBField {
	dbFields := make([]DBField, 0)
	for i := 0; i < structType.NumFields(); i++ {
		f := structType.Field(i)

		// 匿名字段或嵌入字段
		if f.Anonymous() || f.Embedded() {
			dbFields = append(dbFields, processAnonymousOrEmbeddedField(f)...)
			continue
		}

		dbTag := getDBTag(structType.Tag(i))
		// 如果没有
		if len(dbTag) == 0 {
			continue
		}

		dbField := DBField{
			Name:   f.Name(),
			DBName: dbTag,
		}
		dbFields = append(dbFields, dbField)
	}
	return dbFields
}

func getDBTag(tag string) string {
	tag = strings.Trim(tag, "`")
	structTag := reflect.StructTag(tag)
	var dbName = structTag.Get("db")
	if len(dbName) == 0 {
		dbName = structTag.Get("spanner")
	}
	return dbName
}

func generateCode(dirAbsPath string, pkgName string, dbStructs []DBStruct) error {
	// Go 代码模板
	const codeTemplate = `
package {{.Name}}


{{- range .DBStructs}}

type {{.DBName | UpperCaseFirstLetter}}Table struct {
	TableName string
	{{- range .DBFields}}
	{{.Name}} string
	{{- end}}
}

//goland:noinspection GoUnusedGlobalVariable
var {{.DBName | UpperCaseFirstLetter}}TableConst = {{.DBName | UpperCaseFirstLetter}}Table{
		TableName: "{{.DBName | ToSnakeCase}}",
		{{- range .DBFields}}
		{{.Name}}: "{{.DBName}}",
		{{- end}}
	}

{{- end}}
`

	// 创建模板
	tmpl, err := template.New("code").Funcs(template.FuncMap{
		"ToSnakeCase":          toSnakeCase,
		"UpperCaseFirstLetter": UpperCaseFirstLetter,
	}).Parse(codeTemplate)

	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// 生成代码
	param := struct {
		Name      string
		DBStructs []DBStruct
	}{
		Name:      pkgName,
		DBStructs: dbStructs,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, param); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// 格式化代码
	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format code: %w", err)
	}

	// 写入文件
	fileName := fmt.Sprintf("%s/db_gen.go", dirAbsPath)
	if err := ioutil.WriteFile(fileName, formattedCode, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func UpperCaseFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func toSnakeCase(s string) string {
	// 这里简化了蛇形命名的实现，你可能需要一个更复杂的实现来处理特殊情况
	result := make([]byte, 0, len(s)*2)
	for i, c := range s {
		if i > 0 && c >= 'A' && c <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, byte(c|0x20)) // 转换为小写
	}
	return string(result)
}
