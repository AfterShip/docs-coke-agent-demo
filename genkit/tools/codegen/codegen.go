// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package main is a tool to automate the creation of code init function.
// Inspired by `github.com/golang/tools/cmd/stringer`.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/format"
	"go/token"
	"go/types"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

var errCodeDocPrefix = `# 错误码

⚠️⚠️系统错误码列表，由 {{.}}codegen -doc{{.}} 命令生成，不要对此文件做任何更改。⚠️⚠️

## 功能说明

如果返回结果中存在 {{.}}code{{.}} 字段，且 code != Success 则表示调用 API 接口失败。例如：

{{.}}{{.}}{{.}}json
{
  "code": 10001000,
  "message": "Database error",
  "errors": [],
  "request_id": "67575010234d4f9f9adaca7c26e7e709"
}
{{.}}{{.}}{{.}}

上述返回中 {{.}}code{{.}} 表示错误码，{{.}}message{{.}} 表示该错误的具体信息。{{.}}errors{{.}} 表示该错误的 Debug 信息。{{.}}request_id{{.}} 表示该次请求的 id, 可提供给服务开发者用于排查跟踪。
每个错误同时也对应一个 HTTP 状态码，比如上述错误码对应了 HTTP 状态码 500(Internal Server Error)。

## Code 的规则
| 前三位(100) | 中两位(01) | 后三位(000) |
| ---------- | ---- | --------- |
| 系统编码 | 模块编码 | 具体错误码 |


## 错误码列表

本系统支持的错误码列表如下：

| Identifier | Code | HTTP Code | Message |
| ---------- | ---- | --------- | ----------- |
`

var (
	typeNames     = flag.String("type", "", "comma-separated list of type names; must be set")
	output        = flag.String("output", "", "output file name; default srcdir/<type>_gen.go")
	trimprefix    = flag.String("trimprefix", "", "trim the `prefix` from the generated constant names")
	buildTags     = flag.String("tags", "", "comma-separated list of build tags to apply")
	codePackage   = flag.String("codePackage", "code", "package path of code")
	genDocEnabled = flag.Bool("doc", false, "if true only generate error code documentation in markdown format")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of codegen:\n")
	fmt.Fprintf(os.Stderr, "\tcodegen [flags] -type T directries...\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func parserDirectory(dirPath string, tags []string, typeNames []string, codePackagePath string) {
	// Parse the package once.
	g := Generator{
		trimPrefix:      *trimprefix,
		codePackagePath: codePackagePath,
	}

	g.parseDirectory(dirPath, tags)

	// Run generate for each type.
	var src []byte
	for _, typeName := range typeNames {
		if *genDocEnabled {
			g.generateDocs(typeName)
			src = g.buf.Bytes()
		} else {
			g.generate(typeName)
			src = g.format()
		}
	}

	outputFile(src, dirPath)
}

func outputFile(src []byte, dirPath string) {
	// Write to file.
	outputName := *output
	if outputName == "" {
		baseName := "code_gen.go"
		outputName = filepath.Join(dirPath, strings.ToLower(baseName))
	}
	err := ioutil.WriteFile(outputName, src, 0600)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

func getCodePackagePath() string {
	if *codePackage == "" {
		return "code"
	}

	cfg := &packages.Config{
		Mode:  packages.NeedName,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, *codePackage)
	if err != nil {
		log.Fatal(err)
	}

	return pkgs[0].PkgPath
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("codegen: ")

	flag.Usage = Usage
	flag.Parse()
	if len(*typeNames) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	typeNameList := strings.Split(*typeNames, ",")
	var tags []string
	if len(*buildTags) > 0 {
		tags = strings.Split(*buildTags, ",")
	}

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	codePackagePath := getCodePackagePath()

	for _, directory := range args {
		if isDirectory(directory) {
			parserDirectory(directory, tags, typeNameList, codePackagePath)
		} else {
			log.Fatalf("not a directory: %s", directory)
		}
	}
}

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}

	return info.IsDir()
}

// Generator holds the state of the analysis. Primarily used to buffer
// the output for format.Source.
type Generator struct {
	buf               bytes.Buffer // Accumulated output.
	topPackage        *Package
	subPackages       []*Package // Package we are scanning.
	allErrorCodeFiles []*File    // All files, in order encountered.
	trimPrefix        string
	codePackagePath   string
}

// Printf like fmt.Printf, but add the string to g.buf.
func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

// File holds a single parsed file and associated data.
type File struct {
	pkg  *Package  // Package to which this file belongs.
	file *ast.File // Parsed AST.
	// These fields are reset for each type being generated.
	typeName    string  // Name of the constant type.
	values      []Value // Accumulator for constant values of that type.
	trimPrefix  string
	absolutPath string
}

// Package defines options for package.
type Package struct {
	name  string
	path  string
	defs  map[*ast.Ident]types.Object
	files []*File
}

// parseDirectory analyzes the single package constructed from the patterns and tags.
// parseDirectory exits if there is an error.
func (g *Generator) parseDirectory(directoryPath string, tags []string) {
	cfg := &packages.Config{
		Mode:       packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
	}
	pkgs, err := packages.Load(cfg, directoryPath+"/...")
	if err != nil {
		log.Fatal(err)
	}

	directoryName := filepath.Base(directoryPath)

	for _, pkg := range pkgs {
		packageObject := g.parsePackage(pkg)
		var includingCodeFiles = false
		for _, file := range packageObject.files {
			if file.file == nil {
				continue
			}
			baseName := filepath.Base(file.absolutPath)
			if baseName != "code.go" {
				continue
			}
			includingCodeFiles = true
			g.allErrorCodeFiles = append(g.allErrorCodeFiles, file)
		}

		if directoryName == packageObject.name {
			g.topPackage = packageObject
		} else if includingCodeFiles {
			g.subPackages = append(g.subPackages, packageObject)
		}
	}
}

// parsePackage adds a type checked Package and its syntax files to the generator.
func (g *Generator) parsePackage(pkg *packages.Package) *Package {
	currentPackage := &Package{
		name:  pkg.Name,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*File, len(pkg.Syntax)),
		path:  pkg.PkgPath,
	}
	for i, file := range pkg.Syntax {
		absolutPath := pkg.Fset.File(file.Pos()).Name()
		currentPackage.files[i] = &File{
			file:        file,
			pkg:         currentPackage,
			trimPrefix:  g.trimPrefix,
			absolutPath: absolutPath,
		}
	}
	return currentPackage
}

// generate produces the register calls for the named type.
func (g *Generator) generate(typeName string) {
	values := make([]Value, 0, 100)
	for _, file := range g.allErrorCodeFiles {
		// Set the state for this run of the walker.
		file.typeName = typeName
		file.values = nil
		if file.file != nil {
			ast.Inspect(file.file, file.genDecl)
			values = append(values, file.values...)
		}
	}

	if len(values) == 0 {
		log.Fatalf("no values defined for type %s", typeName)
	}

	g.Printf("// Code generated by \"codegen\"; DO NOT EDIT.\n")
	g.Printf("\n")
	g.Printf("package %s", g.topPackage.name)
	g.Printf("\n")

	g.Printf("import \"%s\"\n", g.codePackagePath)

	for _, subPackage := range g.subPackages {
		g.Printf("import %s \"%s\"\n", subPackage.name, subPackage.path)
	}

	g.Printf("\t// init register error codes defines in this source code to `github.com/mingyuans/errors`\n")
	g.Printf("func init() {\n")

	for _, v := range values {
		code, description := v.ParseComment()
		var namePrefix = v.packageName + "."
		if strings.HasSuffix(namePrefix, g.topPackage.name) {
			namePrefix = ""
		}
		g.Printf("\tcode.Register(%s%s, %s, \"%s\")\n", namePrefix, v.originalName, code, description)
	}

	g.Printf("}\n")
}

// generateDocs produces error code markdown document for the named type.
func (g *Generator) generateDocs(typeName string) {
	values := make([]Value, 0, 100)
	for _, file := range g.allErrorCodeFiles {
		// Set the state for this run of the walker.
		file.typeName = typeName
		file.values = nil
		if file.file != nil {
			ast.Inspect(file.file, file.genDecl)
			values = append(values, file.values...)
		}
	}

	if len(values) == 0 {
		log.Fatalf("no values defined for type %s", typeName)
	}

	values = append(values, Value{
		originalName: "Success",
		value:        0,
		comment:      "// Success - 200: OK.",
	})

	// Sort values by value.value
	sort.Slice(values, func(i, j int) bool {
		return values[i].value < values[j].value
	})

	tmpl, _ := template.New("doc").Parse(errCodeDocPrefix)
	var buf bytes.Buffer
	_ = tmpl.Execute(&buf, "`")

	// Generate code that will fail if the constants change value.
	g.Printf("%s", buf.String())
	for _, v := range values {
		code, description := v.ParseComment()
		// g.Printf("\tregister(%s, %s, \"%s\")\n", v.originalName, code, description)
		g.Printf("| %s | %d | %s | %s |\n", v.originalName, v.value, code, description)
	}
	g.Printf("\n")
}

// format returns the gofmt-ed contents of the Generator's buffer.
func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")

		return g.buf.Bytes()
	}

	return src
}

// Value represents a declared constant.
type Value struct {
	comment      string
	originalName string // The name of the constant.
	packageName  string
	name         string // The name with trimmed prefix.
	// The value is stored as a bit pattern alone. The boolean tells us
	// whether to interpret it as an int64 or a uint64; the only place
	// this matters is when sorting.
	// Much of the time the str field is all we need; it is printed
	// by Value.String.
	value  uint64 // Will be converted to int64 when needed.
	signed bool   // Whether the constant is a signed type.
	str    string // The string representation given by the "go/constant" package.
}

func (v *Value) String() string {
	return v.str
}

// ParseComment parse comment to http code and error code description.
func (v *Value) ParseComment() (string, string) {
	reg := regexp.MustCompile(`\w\s*-\s*(\d{3})\s*:\s*([A-Z].*)\s*\.\n*`)
	if !reg.MatchString(v.comment) {
		log.Printf("constant '%s' have wrong comment format, register with 500 as default", v.originalName)

		return "500", "Internal server error"
	}

	groups := reg.FindStringSubmatch(v.comment)
	if len(groups) != 3 {
		return "500", "Internal server error"
	}

	return groups[1], groups[2]
}

// nolint: gocognit
// genDecl processes one declaration clause.
func (f *File) genDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.CONST {
		// We only care about const declarations.
		return true
	}
	// The name of the type of the constants we are declaring.
	// Can change if this is a multi-element declaration.
	typ := ""
	// Loop over the elements of the declaration. Each element is a ValueSpec:
	// a list of names possibly followed by a type, possibly followed by values.
	// If the type and value are both missing, we carry down the type (and value,
	// but the "go/types" package takes care of that).
	for _, spec := range decl.Specs {
		vspec, _ := spec.(*ast.ValueSpec) // Guaranteed to succeed as this is CONST.
		if vspec.Type == nil && len(vspec.Values) > 0 {
			// "X = 1". With no type but a value. If the constant is untyped,
			// skip this vspec and reset the remembered type.
			typ = ""

			// If this is a simple type conversion, remember the type.
			// We don't mind if this is actually a call; a qualified call won't
			// be matched (that will be SelectorExpr, not Ident), and only unusual
			// situations will result in a function call that appears to be
			// a type conversion.
			ce, ok := vspec.Values[0].(*ast.CallExpr)
			if !ok {
				continue
			}
			id, ok := ce.Fun.(*ast.Ident)
			if !ok {
				continue
			}
			typ = id.Name
		}
		if vspec.Type != nil {
			// "X T". We have a type. Remember it.
			ident, ok := vspec.Type.(*ast.Ident)
			if !ok {
				continue
			}
			typ = ident.Name
		}
		if typ != f.typeName {
			// This is not the type we're looking for.
			continue
		}
		// We now have a list of names (from one line of source code) all being
		// declared with the desired type.
		// Grab their names and actual values and store them in f.values.
		for _, name := range vspec.Names {
			if name.Name == "_" {
				continue
			}
			// This dance lets the type checker find the values for us. It's a
			// bit tricky: look up the object declared by the name, find its
			// types.Const, and extract its value.
			obj, ok := f.pkg.defs[name]
			if !ok {
				log.Fatalf("no value for constant %s", name)
			}
			info := obj.Type().Underlying().(*types.Basic).Info()
			if info&types.IsInteger == 0 {
				log.Fatalf("can't handle non-integer constant type %s", typ)
			}
			value := obj.(*types.Const).Val() // Guaranteed to succeed as this is CONST.
			if value.Kind() != constant.Int {
				log.Fatalf("can't happen: constant is not an integer %s", name)
			}
			i64, isInt := constant.Int64Val(value)
			u64, isUint := constant.Uint64Val(value)
			if !isInt && !isUint {
				log.Fatalf("internal error: value of %s is not an integer: %s", name, value.String())
			}
			if !isInt {
				u64 = uint64(i64)
			}
			v := Value{
				originalName: name.Name,
				value:        u64,
				signed:       info&types.IsUnsigned == 0,
				str:          value.String(),
			}
			if vspec.Doc != nil && vspec.Doc.Text() != "" {
				v.comment = vspec.Doc.Text()
			} else if c := vspec.Comment; c != nil && len(c.List) == 1 {
				v.comment = c.Text()
			}

			v.name = strings.TrimPrefix(v.originalName, f.trimPrefix)
			v.packageName = f.file.Name.Name
			f.values = append(f.values, v)
		}
	}

	return false
}
