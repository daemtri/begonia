package helper

// var a = New[struct {
// 	A int `default:"100"`
// }]()
func New[T any]() T {
	var x T
	return x
}
