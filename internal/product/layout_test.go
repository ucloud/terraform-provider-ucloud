package product

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	modulePath        = "github.com/terraform-providers/terraform-provider-ucloud"
	legacyPackage     = modulePath + "/ucloud"
	productAPI        = modulePath + "/internal/product"
	productPathPrefix = modulePath + "/products/"
)

func TestProductPackagesDoNotDependOnLegacyOrOtherProducts(t *testing.T) {
	productsRoot := filepath.Join("..", "..", "products")
	err := filepath.WalkDir(productsRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		relative, err := filepath.Rel(productsRoot, path)
		if err != nil {
			return err
		}
		parts := strings.Split(relative, string(filepath.Separator))
		if len(parts) < 2 {
			t.Errorf("%s must be inside products/<name>", path)
			return nil
		}
		owner := parts[0]

		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Errorf("parse %s: %v", path, err)
			return nil
		}
		for _, imported := range parsed.Imports {
			importPath, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				t.Errorf("parse import %s in %s: %v", imported.Path.Value, path, err)
				continue
			}
			if importPath == legacyPackage || strings.HasPrefix(importPath, legacyPackage+"/") {
				t.Errorf("%s imports legacy provider package %s", path, importPath)
				continue
			}
			if strings.HasPrefix(importPath, productPathPrefix) {
				importedProduct := strings.Split(strings.TrimPrefix(importPath, productPathPrefix), "/")[0]
				if importedProduct != owner {
					t.Errorf("%s imports product %q owned by another package", path, importedProduct)
				}
				continue
			}
			if strings.HasPrefix(importPath, modulePath+"/") && importPath != productAPI {
				t.Errorf("%s imports unsupported core package %s; products may only depend on %s", path, importPath, productAPI)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk product packages: %v", err)
	}
}
