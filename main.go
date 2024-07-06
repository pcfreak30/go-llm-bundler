package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/modfile"
)

type BundleMeta struct {
	D []string            `json:"d,omitempty"` // Dependencies
	I []string            `json:"i,omitempty"` // Imports
	S map[string][]string `json:"s,omitempty"` // Structure (package -> files)
}

type Config struct {
	ProjectDir  string
	OutputFile  string
	ExcludeDirs []string
	IncludeMeta bool
	MinifyLevel int
}

func minifyGoCode(code string, level int) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		return "", err
	}

	// Remove comments
	ast.Inspect(file, func(n ast.Node) bool {
		if _, ok := n.(*ast.CommentGroup); ok {
			return false
		}
		return true
	})

	// Aggressive minification for higher levels
	if level > 1 {
		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Ident:
				// Shorten identifiers (risky, only for very high minification)
				if len(x.Name) > 3 && level > 2 {
					x.Name = x.Name[:3]
				}
			case *ast.GenDecl:
				// Combine var declarations
				if x.Tok == token.VAR && len(x.Specs) > 1 {
					for i := 1; i < len(x.Specs); i++ {
						x.Specs[0].(*ast.ValueSpec).Names = append(x.Specs[0].(*ast.ValueSpec).Names, x.Specs[i].(*ast.ValueSpec).Names...)
						x.Specs[0].(*ast.ValueSpec).Values = append(x.Specs[0].(*ast.ValueSpec).Values, x.Specs[i].(*ast.ValueSpec).Values...)
					}
					x.Specs = x.Specs[:1]
				}
			}
			return true
		})
	}

	// Merge import declarations
	var importSpecs []ast.Spec
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			importSpecs = append(importSpecs, genDecl.Specs...)
		}
	}
	if len(importSpecs) > 0 {
		file.Decls = append([]ast.Decl{&ast.GenDecl{Tok: token.IMPORT, Specs: importSpecs}}, file.Decls...)
	}

	var buf bytes.Buffer
	err = printer.Fprint(&buf, fset, file)
	if err != nil {
		return "", err
	}
	minified := buf.String()

	// Remove unnecessary whitespace
	minified = regexp.MustCompile(`\s+`).ReplaceAllString(minified, " ")
	minified = regexp.MustCompile(`\{\s+`).ReplaceAllString(minified, "{")
	minified = regexp.MustCompile(`\s+\}`).ReplaceAllString(minified, "}")
	minified = regexp.MustCompile(`;\s*`).ReplaceAllString(minified, ";")

	return minified, nil
}

func collectFiles(dir string, config Config) (BundleMeta, map[string]string, error) {
	meta := BundleMeta{
		S: make(map[string][]string),
	}
	files := make(map[string]string)
	var allImports []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(dir, path)

		for _, excludeDir := range config.ExcludeDirs {
			if strings.HasPrefix(relPath, excludeDir) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			minified, err := minifyGoCode(string(content), config.MinifyLevel)
			if err != nil {
				return err
			}

			files[relPath] = minified

			// Extract package and imports
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", minified, parser.ImportsOnly)
			if err == nil {
				pkg := f.Name.Name
				meta.S[pkg] = append(meta.S[pkg], relPath)
				for _, imp := range f.Imports {
					impPath := strings.Trim(imp.Path.Value, "\"")
					if !contains(allImports, impPath) {
						allImports = append(allImports, impPath)
					}
				}
			}
		}

		return nil
	})

	meta.I = allImports
	return meta, files, err
}

func readGoMod(filename string) ([]string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	f, err := modfile.Parse(filename, content, nil)
	if err != nil {
		return nil, err
	}

	var dependencies []string
	for _, req := range f.Require {
		dependencies = append(dependencies, fmt.Sprintf("%s@%s", req.Mod.Path, req.Mod.Version))
	}

	return dependencies, nil
}

func createProjectBundle(config Config) error {
	meta, files, err := collectFiles(config.ProjectDir, config)
	if err != nil {
		return fmt.Errorf("error collecting files: %v", err)
	}

	dependencies, err := readGoMod(filepath.Join(config.ProjectDir, "go.mod"))
	if err != nil {
		return fmt.Errorf("error reading go.mod: %v", err)
	}
	meta.D = dependencies

	if !config.IncludeMeta {
		meta.S = nil // Remove structure info if not needed
	}

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("error creating JSON: %v", err)
	}

	output := string(metaJSON) + "\n"

	for path, content := range files {
		output += fmt.Sprintf("###FILE:%s###\n%s\n", path, content)
	}

	err = os.WriteFile(config.OutputFile, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	fmt.Printf("AI-friendly Go project bundle created: %s\n", config.OutputFile)
	return nil
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func main() {
	config := Config{}

	// Get the absolute path of the current directory
	currentDir, err := filepath.Abs(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Get the base name of the project directory
	defaultProjectName := filepath.Base(currentDir)
	// Sanitize the project name for use in a filename
	defaultProjectName = regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(defaultProjectName, "_")
	defaultOutputFile := fmt.Sprintf("%s_bundle.txt", defaultProjectName)

	flag.StringVar(&config.ProjectDir, "dir", ".", "The root directory of the Go project")
	flag.StringVar(&config.OutputFile, "out", defaultOutputFile, "The output file")
	flag.BoolVar(&config.IncludeMeta, "meta", false, "Include metadata (package structure)")
	flag.IntVar(&config.MinifyLevel, "minify", 1, "Minification level (1-3)")
	excludeDirs := flag.String("exclude", "vendor,testdata", "Comma-separated list of directories to exclude")

	flag.Parse()

	// Update the output filename if a custom project directory is specified
	if config.ProjectDir != "." {
		absProjectDir, err := filepath.Abs(config.ProjectDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting absolute path: %v\n", err)
			os.Exit(1)
		}
		projectName := filepath.Base(absProjectDir)
		projectName = regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(projectName, "_")
		if config.OutputFile == defaultOutputFile {
			config.OutputFile = fmt.Sprintf("%s_bundle.txt", projectName)
		}
	}

	config.ExcludeDirs = strings.Split(*excludeDirs, ",")

	if err := createProjectBundle(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
