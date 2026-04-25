package dispatch

type Task func()

var Actions = make(chan Task, 100)
