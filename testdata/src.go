package main

type MyStruct struct {
	Field *MyStruct
}

func main() {
	x := MyStruct{Field: &MyStruct{}}
	// This should be transformed to println(x.GetField())
	println(x.Field) // testing inline comment
	// This should be transformed to x.SetField(&MyStruct{})
	x.Field = &MyStruct{}

	// This should be transformed to foo := x.GetField().GetField()
	foo := x.Field.Field
	_ = foo
}
