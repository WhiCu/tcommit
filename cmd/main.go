package main

import (
	"io"
	"os"

	"github.com/WhiCu/TCommit/internal/core/git"
	"github.com/WhiCu/TCommit/internal/core/template"
)

func main() {
	cont, err := os.Open(".tcommit")
	if err != nil {
		panic(err)
	}
	values := map[string]string{
		"hime": "HIME_VALUE",
		"get":  "ui",
	}
	read := template.New(cont, values)
	res, err := io.ReadAll(read.Parse())
	if err != nil {
		panic(err)
	}
	err = git.Commit(string(res))
	if err != nil {
		panic(err)
	}
}
