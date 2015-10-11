// Transforms a Golang source file that uses AES encryption (with fixed keys) for some task into a functionally
// equivalent one where the fixed keys can't be extracted (with white-box AES).
//
// To run the program: go run transform.go -in input.go -out output.go
// All calls to aes.NewCipher with a 16-byte long []byte as the key argument will be transformed to hide the key.
package main

import (
	"crypto/rand"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	"github.com/OpenWhiteBox/AES/constructions/chow"
)

type KeyAndPos struct {
	Key []byte
	Pos token.Pos
}

const (
	target      = "\"crypto/aes\""
	replacement = "\"github.com/OpenWhiteBox/AES/constructions/chow\""
)

var (
	in  = flag.String("in", "input.go", "Input source file.")
	out = flag.String("out", "output.go", "Output source file.")
)

func main() {
	flag.Parse()

	src, err := ioutil.ReadFile(*in)
	if err != nil {
		log.Fatal(err)
	}

	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "", src, parser.AllErrors)

	// Find where "crypto/aes" is imported.
	var importSpec *ast.ImportSpec

	for _, imp := range f.Imports {
		if imp.Path.Value == target {
			importSpec = imp
			break
		}
	}

	if importSpec == nil {
		log.Fatalf("File doesn't import %v!", target)
	}

	// Name the import explicitly if it's not already named. Change the import path to a white-box construction.
	if importSpec.Name == nil {
		importSpec.Name = &ast.Ident{0, "aes", nil}
	} else if importSpec.Name.Name == "." {
		log.Fatalf("Can't transform a file's encryption keys if %v is imported into the local scope!", target)
	}

	importSpec.Path.Value = replacement

	// Find everywhere we initialize a block cipher with an explicit key and replace it with a white-box.
	ast.Inspect(f, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.CallExpr:
			callExpr := n.(*ast.CallExpr)

			if IsCallToEncrypt(importSpec.Name.Name, callExpr) {
				key, ok := ExtractKey(callExpr.Args[0]) // aes.NewCipher has only one argument.

				if ok {
					TransformCallToEncrypt(callExpr, key)
					return false
				} else {
					log.Printf("Found encryption call at %v, but couldn't extract the key!", fset.Position(n.Pos()))
					return true
				}
			}
		}

		return true
	})

	dst, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}

	printer.Fprint(dst, fset, f)
}

func IsCallToEncrypt(pkgName string, callExpr *ast.CallExpr) bool {
	selector, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "NewCipher" || selector.Sel.Obj != nil {
		return false
	}

	x, ok := selector.X.(*ast.Ident)
	if !ok || x.Name != pkgName || x.Obj != nil { // Imported packages always have a nil object.
		return false
	}

	return true
}

func TransformCallToEncrypt(callExpr *ast.CallExpr, key []byte) {
	callExpr.Fun.(*ast.SelectorExpr).Sel.Name = "Parse"
	callExpr.Args[0] = &ast.CompositeLit{
		Type: &ast.ArrayType{
			Lbrack: 0,
			Len:    nil,
			Elt:    &ast.Ident{0, "byte", nil},
		},
		Lbrace: 0,
		Elts:   make([]ast.Expr, len(key)),
		Rbrace: 0,
	}

	for i, elem := range key {
		callExpr.Args[0].(*ast.CompositeLit).Elts[i] = &ast.BasicLit{
			ValuePos: 0,
			Kind:     token.INT,
			Value:    strconv.Itoa(int(elem)),
		}
	}
}

func ExtractKey(cand ast.Expr) ([]byte, bool) {
	// Verify expression contains a byte slice.
	compositeLit, ok := cand.(*ast.CompositeLit)
	if !ok {
		return nil, false
	}

	arrayType, ok := compositeLit.Type.(*ast.ArrayType)
	if !ok {
		return nil, false
	}

	elt, ok := arrayType.Elt.(*ast.Ident)
	if !ok || elt.Name != "byte" || elt.Obj != nil {
		return nil, false
	}

	// Extract raw key.
	key := []byte{}

	for _, e := range compositeLit.Elts {
		basicLit, ok := e.(*ast.BasicLit)
		if !ok || basicLit.Kind != token.INT {
			return nil, false
		}

		num, err := strconv.ParseUint(basicLit.Value, 0, 8)
		if err != nil {
			return nil, false
		}

		key = append(key, byte(num))
	}

	if len(key) != 16 {
		return nil, false
	}

	// Mask key.
	seed := make([]byte, 16)
	rand.Read(seed)

	constr, _, _ := chow.GenerateKeys(key, seed, chow.SameMasks(chow.IdentityMask))

	return constr.Serialize(), true
}
