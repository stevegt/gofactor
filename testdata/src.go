package main

type MyStruct struct {
	Field int
}

func main() {
	x := MyStruct{Field: 10}
	// This should be transformed to println(x.GetField())
	println(x.Field) // testing inline comment
	// This should be transformed to x.SetField(20)
	x.Field = 20
}
