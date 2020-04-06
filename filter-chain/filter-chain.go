package filter_chain

import "github.com/fpablos/web-proxy-service/couchbase"

var db = couchbase.GetInstance()

// Executer is an interface for filters.
type Executer interface {
	Execute(*Chain, ...interface{}) bool
}

// Inline is a type for adding filters as anonymous functions.
//    chain.AddFilter(&filterchain.Inline{func(chain *filterchain.Chain, args ...interface{}) error {
//        err := chain.Next(args)
//        return err
//    }})
type Inline struct {
	Handler func(*Chain, ...interface{}) bool
}

// Execute runs the inlined handler.
func (filter *Inline) Execute(chain *Chain, args ...interface{}) bool {
	return filter.Handler(chain, args...)
}

// Chain is the main type.
type Chain struct {
	pos int
	filters []Executer
}

// New creates new chain.
func New() *Chain {
	return &Chain{}
}

// AddFilter adds a filter to the chain.
func (chain *Chain) AddFilter(filter Executer) *Chain {
	chain.filters = append(chain.filters, filter)
	return chain
}

// Execute starts executing filters in the chain.
func (chain *Chain) Execute(args ...interface{}) bool {
	pos := chain.pos
	if pos < len(chain.filters) {
		chain.pos++
		if goToNextChain := chain.filters[pos].Execute(chain, args...); goToNextChain == true {
			return true
		}
	}
	return false
}

// Next executes the next filter in the chain.
func (chain *Chain) Next(args ... interface{}) bool {
	return chain.Execute(args...)
}

// Rewind rewinds the chain, so it can be run again.
func (chain *Chain) Rewind() {
	chain.pos = 0
}