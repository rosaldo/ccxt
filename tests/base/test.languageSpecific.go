package base

import ccxt "github.com/rosaldo/ccxt/v5"

func TestLanguageSpecificAsync() <-chan ccxt.AsyncResult[any] {
	TestFutures()
	TestStructs()
	TestOptionTypes()

	ch := make(chan ccxt.AsyncResult[any])
	close(ch)
	return ch
}
