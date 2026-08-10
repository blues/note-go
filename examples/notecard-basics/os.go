package main

import (
	"github.com/blues/note-go/notecard"
)

var (
	port          string
	speed         int
	cardInterface = "serial"
)

func init() {
	_, port, speed = notecard.Defaults()
}
