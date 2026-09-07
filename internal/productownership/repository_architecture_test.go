package productownership_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productcatalog"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownership"
)

const modulePath = "github.com/terraform-providers/terraform-provider-ucloud"

func TestRepositoryProductArchitecture(t *testing.T) {
	root := repositoryRoot(t)
	policy := loadRepositoryPolicy(t, root)
	wantProducts := productcatalog.Names()

	t.Run("ownership policy matches catalog", func(t *testing.T) {
		got := sortedProductNames(policy.Products)
		assertStringSlicesEqual(t, got, wantProducts)
	})

	t.Run("product directories match catalog", func(t *testing.T) {
		got := productDirectories(t, root)
		assertStringSlicesEqual(t, got, wantProducts)
	})

	t.Run("provider uses catalog bindings", func(t *testing.T) {
		assertProviderUsesCatalog(t, filepath.Join(root, "ucloud", "provider.go"))
	})

	t.Run("provider starts with empty registration maps", func(t *testing.T) {
		assertEmptyProviderMaps(t, filepath.Join(root, "ucloud", "provider.go"))
	})

	t.Run("every product owns acceptance tests", func(t *testing.T) {
		for _, name := range wantProducts {
			name := name
			t.Run(name, func(t *testing.T) {
				baseline, tracked := productcatalog.Baseline(name)
				if !tracked {
					t.Fatalf("products/%s has no compatibility test baseline", name)
				}
				if count := countTestFunctions(t, filepath.Join(root, "products", name), "TestAcc"); count < baseline.AcceptanceTests {
					t.Fatalf("products/%s has %d TestAcc functions; compatibility baseline requires at least %d", name, count, baseline.AcceptanceTests)
				}
				if count := countTestFunctionsWithSuffix(t, filepath.Join(root, "products", name), "TestAcc", "_import"); count < baseline.ImportTests {
					t.Fatalf("products/%s has %d import tests; compatibility baseline requires at least %d", name, count, baseline.ImportTests)
				}
			})
		}
	})

	t.Run("core contains no product implementations or acceptance tests", func(t *testing.T) {
		assertThinCore(t, root)
	})

	t.Run("product production code is isolated", func(t *testing.T) {
		assertProductImportsAreIsolated(t, root)
	})
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve repository architecture test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func loadRepositoryPolicy(t *testing.T, root string) *productownership.Policy {
	t.Helper()
	file, err := os.Open(filepath.Join(root, ".github", "product-owners.json"))
	if err != nil {
		t.Fatalf("open product ownership policy: %v", err)
	}
	defer file.Close()

	policy, err := productownership.Load(file)
	if err != nil {
		t.Fatalf("load product ownership policy: %v", err)
	}
	return policy
}

func sortedProductNames(products map[string]productownership.Product) []string {
	names := make([]string, 0, len(products))
	for name := range products {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func productDirectories(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "products"))
	if err != nil {
		t.Fatalf("read products directory: %v", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names
}

func assertProviderUsesCatalog(t *testing.T, filename string) {
	t.Helper()
	file := parseGoFile(t, filename)
	directBindings := 0
	catalogCalls := 0
	catalogRegistrations := 0
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}
		if pkg.Name == "product" && selector.Sel.Name == "Bind" {
			directBindings++
		}
		if pkg.Name == "productcatalog" && selector.Sel.Name == "Bindings" {
			catalogCalls++
		}
		if pkg.Name == "product" && selector.Sel.Name == "MustRegister" &&
			len(call.Args) == 2 && call.Ellipsis.IsValid() &&
			isPackageCall(call.Args[1], "productcatalog", "Bindings") {
			catalogRegistrations++
		}
		return true
	})
	if directBindings != 0 {
		t.Fatalf("provider contains %d direct product.Bind calls; bindings must come from the catalog", directBindings)
	}
	if catalogCalls != 1 || catalogRegistrations != 1 {
		t.Fatalf("provider catalog calls = %d and registrations = %d; want one productcatalog.Bindings registration", catalogCalls, catalogRegistrations)
	}
}

func isPackageCall(expression ast.Expr, packageName, functionName string) bool {
	call, ok := expression.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != functionName {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	return ok && pkg.Name == packageName
}

func assertEmptyProviderMaps(t *testing.T, filename string) {
	t.Helper()
	file := parseGoFile(t, filename)
	foundProvider := false
	foundMaps := map[string]bool{
		"DataSourcesMap": false,
		"ResourcesMap":   false,
	}

	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok || !isSelectorType(literal.Type, "schema", "Provider") {
			return true
		}
		foundProvider = true
		for _, element := range literal.Elts {
			field, ok := element.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := field.Key.(*ast.Ident)
			if !ok {
				continue
			}
			if _, tracked := foundMaps[key.Name]; !tracked {
				continue
			}
			value, ok := field.Value.(*ast.CompositeLit)
			if !ok || len(value.Elts) != 0 {
				t.Fatalf("schema.Provider.%s must start as an empty map", key.Name)
			}
			foundMaps[key.Name] = true
		}
		return true
	})

	if !foundProvider {
		t.Fatal("schema.Provider literal not found")
	}
	for name, found := range foundMaps {
		if !found {
			t.Fatalf("schema.Provider.%s empty map not found", name)
		}
	}
}

func assertThinCore(t *testing.T, root string) {
	t.Helper()
	coreDir := filepath.Join(root, "ucloud")
	forbiddenPrefixes := []string{
		"data_source_ucloud_",
		"resource_ucloud_",
		"service_ucloud_",
	}

	entries, err := os.ReadDir(coreDir)
	if err != nil {
		t.Fatalf("read ucloud directory: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		for _, prefix := range forbiddenPrefixes {
			if strings.HasPrefix(entry.Name(), prefix) {
				t.Errorf("legacy product implementation remains in ucloud: %s", entry.Name())
			}
		}
	}
	if count := countTestFunctions(t, coreDir, "TestAcc"); count != 0 {
		t.Errorf("ucloud contains %d TestAcc functions; product acceptance tests must live with their product", count)
	}

	assertCoreSDKImports(t, coreDir)
	assertCoreClientFields(t, filepath.Join(coreDir, "client.go"))
}

func assertCoreSDKImports(t *testing.T, coreDir string) {
	t.Helper()
	const servicesPrefix = "github.com/ucloud/ucloud-sdk-go/services/"
	err := filepath.WalkDir(coreDir, func(filename string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(filename, ".go") || strings.HasSuffix(filename, "_test.go") {
			return nil
		}
		file := parseGoFile(t, filename)
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("decode import in %s: %v", filename, err)
			}
			if strings.HasPrefix(importPath, servicesPrefix) && importPath != servicesPrefix+"sts" {
				t.Errorf("core imports product SDK %q in %s", importPath, filepath.Base(filename))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk ucloud directory: %v", err)
	}
}

func assertCoreClientFields(t *testing.T, filename string) {
	t.Helper()
	want := []string{"config", "credential", "productClients", "projectId", "region", "requestHandlers"}
	file := parseGoFile(t, filename)
	var got []string
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, spec := range general.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "UCloudClient" {
				continue
			}
			structure, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				t.Fatal("UCloudClient is not a struct")
			}
			for _, field := range structure.Fields.List {
				for _, name := range field.Names {
					got = append(got, name.Name)
				}
			}
		}
	}
	sort.Strings(got)
	assertStringSlicesEqual(t, got, want)
}

func assertProductImportsAreIsolated(t *testing.T, root string) {
	t.Helper()
	productsDir := filepath.Join(root, "products")
	err := filepath.WalkDir(productsDir, func(filename string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(filename, ".go") || strings.HasSuffix(filename, "_test.go") {
			return nil
		}
		file := parseGoFile(t, filename)
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("decode import in %s: %v", filename, err)
			}
			if importPath == modulePath+"/ucloud" || strings.HasPrefix(importPath, modulePath+"/products/") {
				t.Errorf("product production code %s imports forbidden package %q", filename, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk products directory: %v", err)
	}
}

func countTestFunctions(t *testing.T, directory, prefix string) int {
	return countTestFunctionsWithSuffix(t, directory, prefix, "")
}

func countTestFunctionsWithSuffix(t *testing.T, directory, prefix, suffix string) int {
	t.Helper()
	count := 0
	err := filepath.WalkDir(directory, func(filename string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(filename, "_test.go") {
			return nil
		}
		file := parseGoFile(t, filename)
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if ok && strings.HasPrefix(function.Name.Name, prefix) && strings.HasSuffix(function.Name.Name, suffix) {
				count++
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", directory, err)
	}
	return count
}

func parseGoFile(t *testing.T, filename string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}
	return file
}

func isSelectorType(expression ast.Expr, packageName, typeName string) bool {
	selector, ok := expression.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != typeName {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	return ok && pkg.Name == packageName
}

func assertStringSlicesEqual(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("got %v, want %v", got, want)
	}
}
