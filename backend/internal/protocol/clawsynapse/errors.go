package clawsynapse

import "errors"

var (
	errEmptyEnvelope = errors.New("clawsynapse: empty envelope")
	errEmptyMessage  = errors.New("clawsynapse: empty message payload")
)
