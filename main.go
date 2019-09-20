package main

import (
	"context"
	"encoding/json"
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

	r1 := rego.New(rego.ParsedQuery(q),
		rego.Compiler(compiler),
		rego.Store(store))

	pq1, err := r1.Partial(ctx)

	err = writePartialQueriesToFile(pq1)
	if err != nil {
		fmt.Printf("failed to write partial queries to file: %v\n", err)
		os.Exit(1)
	}

	// FIXME: this pq2 isn't the same as pq1 and causes failures
	pq2, err := readPartialQueriesFromFile()
	if err != nil {
		fmt.Printf("failed to read partial queries from file: %v\n", err)
		os.Exit(1)
	}

	mods2 := map[string]*ast.Module{
		"foo": pq2.Support[0],
	}

	compiler2, err := initCompiler(mods2)
	if err != nil {
		fmt.Printf("failed to compile 2nd: %v\n", err)
		os.Exit(1)
	}

	r2 := rego.New(rego.ParsedQuery(pq2.Queries[0]),
		rego.Compiler(compiler2),
		rego.Store(inmem.New()),
		rego.ParsedInput(input),
	)

	rs2, err := r2.Eval(ctx)
	if err != nil {
		fmt.Printf("failed to evaluate 2nd: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("rs2: %v\n", rs2)

}

func readPartialQueriesFromFile() (*rego.PartialQueries, error) {
	pqb, err := ioutil.ReadFile("partial_queries")
	if err != nil {
		return nil, err
	}

	pq := decodePartialQueries(pqb)
	return pq, nil

}

func writePartialQueriesToFile(pq *rego.PartialQueries) error {
	b := encodePartialQueries(pq)
	file, err := os.Create(`partial_queries`)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(b)
	if err != nil {
		return err
	}
	return nil
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

func encodePartialQueries(pq *rego.PartialQueries) []byte {
	buf, err := json.Marshal(pq)
	if err != nil {
		panic(fmt.Sprintf("error encoding %v to bytes: %v", pq, err))
	}
	return buf
}

func decodePartialQueries(buf []byte) *rego.PartialQueries {
	var pq rego.PartialQueries
	err := json.Unmarshal(buf, &pq)
	if err != nil {
		panic(fmt.Sprintf("error decoding from bytes: %v", err))
	}
	return &pq
}
