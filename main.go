package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/docopt/docopt-go"
)

const usage = `docopt-ver
	
docopt version setter.

Tool will search for docopt.Parse call in specified file and replace 4-th
argument, which is program version, by specified string value. If original
version string contains program name, like "shadowc 1.1", then only version
part (e.g. "1.1") will be replaced.

Specified file will be changed in-place.
Usage:
    $0 -h | --help
    $0 <file-name> <new-version>

Options:
    -h --help  Show this help.
`

type VersionReplacement struct {
	pos    int
	length int
	value  string
}

func main() {
	args, err := docopt.Parse(
		strings.Replace(usage, "$0", os.Args[0], -1),
		nil, true, "1.0", false,
	)

	var (
		fileName   = args["<file-name>"].(string)
		newVersion = args["<new-version>"].(string)
	)

	if err != nil {
		panic(err)
	}
	fileSet := token.NewFileSet()

	rootNode, err := parser.ParseFile(
		fileSet, fileName, nil, 0,
	)
	if err != nil {
		log.Fatal(err)
	}

	replacement, err := getVersionReplacement(rootNode, newVersion)
	if err != nil {
		log.Fatal(err)
	}

	oldContents, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	newContents := strings.Join(
		[]string{
			string(oldContents[:replacement.pos-1]),
			replacement.value,
			string(oldContents[replacement.pos+replacement.length-1:]),
		},
		"",
	)

	err = ioutil.WriteFile(fileName, []byte(newContents), 0)
	if err != nil {
		log.Fatal(err)
	}
}

func getVersionReplacement(
	rootNode ast.Node, newVersion string,
) (*VersionReplacement, error) {
	var versionReplacement *VersionReplacement
	var err error

	ast.Inspect(rootNode, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
				if pkg, ok := selector.X.(*ast.Ident); ok {
					if pkg.Name == "docopt" && selector.Sel.Name == "Parse" {
						versionReplacement, err = parseDocoptCall(
							node.(*ast.CallExpr), newVersion,
						)

						return false
					}
				}
			}
		}

		return true
	})

	return versionReplacement, err
}

func parseDocoptCall(
	docoptCall *ast.CallExpr, newVersion string,
) (*VersionReplacement, error) {
	if len(docoptCall.Args) < 4 {
		return nil, fmt.Errorf(
			"unexpected number of docopt.Parse arguments, "+
				"at least 4 required, %d found",
			len(docoptCall.Args),
		)
	}

	versionString := docoptCall.Args[3]
	if literal, ok := versionString.(*ast.BasicLit); !ok {
		return nil, fmt.Errorf(
			"expected simple literal in 4-th argument of docopt.Parse, "+
				"found: %#v",
			versionString,
		)
	} else {
		oldValue := literal.Value

		matches := regexp.MustCompile(
			"([\"`'])(.* )?(.*)[\"`']",
		).FindStringSubmatch(
			oldValue,
		)

		newValue := matches[1] + matches[2] + newVersion + matches[1]

		return &VersionReplacement{
			pos:    int(literal.Pos()),
			length: len(oldValue),
			value:  newValue,
		}, nil
	}
}
