package cdc

import (
	"fmt"
	"sync"

	"github.com/vkalekis/companies/pkg/model"
)

type Op string

const (
	Op_Create = "c"
	Op_Update = "u"
	Op_Delete = "d"
)

type Operation struct {
	Before, After *model.Company
	Op            Op
}

func (op Operation) String() string {
	return fmt.Sprintf("[Op=%s] Before: %+v, After: %+v", op.Op, op.Before, op.After)
}

type Operator interface {
	LogCDCOperation(Operation)
}

type Checker struct {
	ch       chan Operation
	operator Operator

	wg     sync.WaitGroup
	quitCh chan struct{}
}

func NewChecker(operator Operator) *Checker {
	return &Checker{
		ch:       make(chan Operation),
		operator: operator,
		wg:       sync.WaitGroup{},
		quitCh:   make(chan struct{}),
	}
}

func (c *Checker) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-c.quitCh:
				return
			case op := <-c.ch:
				c.operator.LogCDCOperation(op)
			}
		}

	}()
}

func (c *Checker) Stop() {
	close(c.quitCh)
	c.wg.Wait()
}

func (c *Checker) Register(op Operation) {
	c.ch <- op
}
