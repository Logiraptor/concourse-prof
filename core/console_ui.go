package main

import "fmt"

type consoleUi struct {
}

func (consoleUi) ShowError(err error) {
	fmt.Println(err)
}
func (consoleUi) SetBarTotal(name string, total int) {
	fmt.Println(name, "has this many total:", total)
}
func (consoleUi) IncrementBar(name string, value int) {
	fmt.Println(name, "just processed this many:", value)
}
