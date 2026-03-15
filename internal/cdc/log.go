package cdc

import "log"

type LogCDCOperator struct{}

func (operator *LogCDCOperator) LogCDCOperation(op Operation) {
	log.Printf("%s", op)
}
