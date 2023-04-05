//
// MIT License
//
// # Copyright (c) 2021 Robert boobusy
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package fastraft

import (
	"fmt"
	"log"
)

type Logger interface {
	Debug(...any)
	Info(...any)
	Warn(...any)
	Error(...any)
	Panic(...any)
}

type logger struct {
	*log.Logger
}

func DefaultLogger() Logger {
	l := new(logger)
	l.Logger = log.Default()
	return l
}

func (l *logger) Debug(args ...any) {
	l.Print("debug ", " ", fmt.Sprintln(args...))
}

func (l *logger) Info(args ...any) {
	l.Print("info", " ", fmt.Sprintln(args...))
}

func (l *logger) Warn(args ...any) {
	l.Print("warn ", " ", fmt.Sprintln(args...))
}

func (l *logger) Error(args ...any) {
	l.Print("error", " ", fmt.Sprintln(args...))
}

// Panic exit process
func (l *logger) Panic(args ...any) {
	l.Print("panic", " ", fmt.Sprintln(args...))
	panic("")
}
