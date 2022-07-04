package utils

import "context"

type ErrorMonad interface {
	Then(func()) ErrorMonad
}

type errorMonad struct {
	ctx context.Context
}

func (m errorMonad) Then(fn func()) ErrorMonad {
	if !HasError(m.ctx) {
		fn()
	}
	return m
}

func StopOnError(ctx context.Context) ErrorMonad {
	return errorMonad{ctx: ctx}
}
