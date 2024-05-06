# gofactor

Currently supports replacing direct field references with
getter/setter methods by parsing and rewriting the AST; will add other
functionality as needed.  I'm using vim-go, GoLand, and GPT-4 for
other refactoring right now.

I use e.g. GoLand or GPT to generate the getter/setter methods, then
run gofactor to replace the direct field references.

