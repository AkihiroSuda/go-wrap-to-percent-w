package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"os/exec"
	"strconv"

	// use "github.com/pkg/errors" for demonstration purpose
	"github.com/pkg/errors"
)

func main() {
	if err := xmain(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func xmain() error {
	flagW := flag.Bool("w", false, "write result to (source) file instead of stdout")
	flagGofmt := flag.Bool("gofmt", true, "run gofmt after conversion")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s [flags] [file ...]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}
	for _, f := range args {
		var transformed bytes.Buffer
		if err := transformFile(&transformed, f); err != nil {
			return errors.Wrapf(err, "failed to transform file %q", f)
		}
		var res *bytes.Buffer = &transformed
		if *flagGofmt {
			var formatted bytes.Buffer
			if err := gofmt(&formatted, &transformed); err != nil {
				return errors.Wrap(err, "failed to gofmt")
			}
			res = &formatted
		}
		if *flagW {
			if err := os.WriteFile(f, res.Bytes(), 0644); err != nil {
				return errors.Wrapf(err, "failed to write the result to %q", f)
			}
		} else {
			if _, err := fmt.Print(res.String()); err != nil {
				return errors.Wrap(err, "failed to print the result to stdout")
			}
		}
	}
	return nil
}

func gofmt(w io.Writer, r io.Reader) error {
	cmd := exec.Command("gofmt")
	cmd.Stdin = r
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func transformFile(w io.Writer, fName string) error {
	fSet := token.NewFileSet()
	astFile, err := parser.ParseFile(fSet, fName, nil,
		parser.ParseComments|parser.AllErrors)
	if err != nil {
		return errors.Wrapf(err, "failed to parse %q", fName)
	}
	if err := transformAST(fSet, astFile); err != nil {
		return errors.Wrapf(err, "failed to transform %q", fName)
	}
	if err := printer.Fprint(w, fSet, astFile); err != nil {
		return errors.Wrap(err, "failed to print the transformed code")
	}
	return nil
}

const githubComPkgErrors = "github.com/pkg/errors"

func transformAST(fSet *token.FileSet, astFile *ast.File) error {
	v := &visitor{
		githubComPkgErrorsLocalName: "errors",
	}
	var (
		githubComPkgErrorsFound bool
		fmtFound                bool
	)
	for _, im := range astFile.Imports {
		// im.Path.Value is quoted with double-quote symbols
		if im != nil && im.Path != nil {
			switch unquote(im.Path.Value) {
			case githubComPkgErrors:
				githubComPkgErrorsFound = true
				if im.Name != nil && im.Name.Name != "" {
					v.githubComPkgErrorsLocalName = unquote(im.Name.Name)
				}
			case "fmt":
				fmtFound = true
			}
		}
	}
	if !githubComPkgErrorsFound {
		// "github.com/pkg/errors" is not used, nothing to do.
		return nil
	}
	ast.Walk(v, astFile)
	if v.fmtNeeded && !fmtFound {
		prependImport(astFile, "fmt")
	}
	if !v.githubComPkgErrorsNeeded {
		removeImport(astFile, githubComPkgErrors)
	}
	if v.stdErrorsNeeded {
		prependImport(astFile, "errors")
	}
	return nil
}

func removeImport(astFile *ast.File, importPath string) {
	importsIdx := -1
	for i, im := range astFile.Imports {
		if im != nil && im.Path != nil && unquote(im.Path.Value) == importPath {
			importsIdx = i
		}
	}
	if importsIdx > 0 {
		astFile.Imports = append(astFile.Imports[:importsIdx], astFile.Imports[importsIdx+1:]...)
	}
	for _, decl := range astFile.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			switch d.Tok {
			case token.IMPORT:
				spIdx := -1
				for i, sp := range d.Specs {
					switch im := sp.(type) {
					case *ast.ImportSpec:
						if im != nil && im.Path != nil && unquote(im.Path.Value) == importPath {
							spIdx = i
						}
					}
				}
				if spIdx > 0 {
					d.Specs = append(d.Specs[:spIdx], d.Specs[spIdx+1:]...)
				}
			}
		}
	}
}

func prependImport(astFile *ast.File, importPath string) {
	importSpec := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(importPath),
		},
	}
	astFile.Imports = append([]*ast.ImportSpec{importSpec}, astFile.Imports...)
	for _, decl := range astFile.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			switch d.Tok {
			case token.IMPORT:
				d.Specs = append([]ast.Spec{importSpec}, d.Specs...)
			}
		}
	}
}
