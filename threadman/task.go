package threadman

type Task struct {
	ID     int
	Func   func() interface{}
	Result interface{}
}
