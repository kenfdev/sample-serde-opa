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

// PartialQuery is a string representation of the rego.PartialQueries
type PartialQuery struct {
	Query   string
	Support string
}

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
		rego.Store(store),
		rego.ParsedInput(input),
	)
	rs1, err := r1.Eval(ctx)
	if err != nil {
		fmt.Printf("failed to evaluate 2nd: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("1st ResultSet: %v\n", rs1)

	pq1, err := r1.Partial(ctx)

	err = writePartialQueriesToFile(pq1)
	if err != nil {
		fmt.Printf("failed to write partial queries to file: %v\n", err)
		os.Exit(1)
	}

	pqs, err := readPartialQueriesFromFile()
	if err != nil {
		fmt.Printf("failed to read partial queries from file: %v\n", err)
		os.Exit(1)
	}

	mods2 := map[string]*ast.Module{
		"foo": ast.MustParseModule(pqs.Support),
	}

	compiler2, err := initCompiler(mods2)
	if err != nil {
		fmt.Printf("failed to compile 2nd: %v\n", err)
		os.Exit(1)
	}

	r2 := rego.New(rego.Query(pqs.Query),
		rego.Compiler(compiler2),
		rego.Store(inmem.New()),
		rego.ParsedInput(input),
	)

	rs2, err := r2.Eval(ctx)
	if err != nil {
		fmt.Printf("failed to evaluate 2nd: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("2nd ResultSet: %v\n", rs2)
}

func readPartialQueriesFromFile() (*PartialQuery, error) {
	pqb, err := ioutil.ReadFile("partial_queries")
	if err != nil {
		return nil, err
	}

	pq := decodePartialQueries(pqb)
	return pq, nil
}

func writePartialQueriesToFile(pq *rego.PartialQueries) error {
	pqs := PartialQuery{
		Query:   pq.Queries[0].String(),
		Support: pq.Support[0].String(),
	}
	b := encodePartialQueries(pqs)
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

func encodePartialQueries(pq PartialQuery) []byte {
	buf, err := json.Marshal(pq)
	if err != nil {
		panic(fmt.Sprintf("error encoding %v to bytes: %v", pq, err))
	}
	return buf
}

func decodePartialQueries(buf []byte) *PartialQuery {
	var pq PartialQuery
	err := json.Unmarshal(buf, &pq)
	if err != nil {
		panic(fmt.Sprintf("error decoding from bytes: %v", err))
	}
	return &pq
}
