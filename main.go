package main

import (
	"context"
	"fmt"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage/inmem"
	"io/ioutil"
	"os"
)

var authzPolicyPath = "policy/authz.rego"
var authzQuery = "data.authz.allow"

func main() {
	ctx := context.Background()

	input := ast.NewObject(
		[2]*ast.Term{ast.NewTerm(ast.String("user")), ast.NewTerm(ast.String("alice"))},
	)

	mods, err := initModules()
	if err != nil {
		fmt.Printf("failed to init modules: %v\n", err)
		os.Exit(1)
	}

	compiler, err := initCompiler(mods)
	if err != nil {
		fmt.Printf("failed to compile modules: %v\n", err)
		os.Exit(1)
	}

	store := inmem.NewFromObject(map[string]interface{}{
		"policies": map[string]interface{}{
			"alice": map[string]string{
				"effect": "allow",
			},
		},
	})

	q, err := initQuery()
	if err != nil {
		fmt.Printf("failed to parse query: %v\n", err)
		os.Exit(1)
	}

	r := rego.New(rego.ParsedQuery(q),
		rego.Compiler(compiler),
		rego.Store(store))

	pq, err := r.Partial(ctx)
	// pr, err := r.PrepareForPartial(ctx)

	mods2 := map[string]*ast.Module{
		"hoge": pq.Support[0],
	}

	compiler2, err := initCompiler(mods2)
	if err != nil {
		fmt.Printf("failed to compile 2nd: %v\n", err)
		os.Exit(1)
	}

	r2 := rego.New(rego.ParsedQuery(pq.Queries[0]),
		rego.Compiler(compiler2),
		rego.Store(inmem.New()),
		rego.ParsedInput(input),
	)
	pq2, err := r2.Partial(ctx)
	if err != nil {
		fmt.Printf("failed to partial eval 2nd: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("pq1: %v\n", pq)
	fmt.Printf("pq2: %v\n", pq2)

	rs2, err := r2.Eval(ctx)
	if err != nil {
		fmt.Printf("failed to evaluate 2nd: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("rs2: %v\n", rs2)

}

func initModules() (map[string]*ast.Module, error) {
	b, err := ioutil.ReadFile("policy/authz.rego")
	if err != nil {
		return nil, err
	}

	m, err := ast.ParseModule(authzPolicyPath, string(b))
	if err != nil {
		return nil, err
	}

	mods := map[string]*ast.Module{
		authzPolicyPath: m,
	}

	return mods, nil
}

func initCompiler(mods map[string]*ast.Module) (*ast.Compiler, error) {
	compiler := ast.NewCompiler()
	compiler.Compile(mods)
	if compiler.Failed() {
		return nil, compiler.Errors
	}

	return compiler, nil
}

func initQuery() (ast.Body, error) {
	parsedQuery, err := ast.ParseBody(authzQuery)
	if err != nil {
		return nil, err
	}
	return parsedQuery, nil
}
