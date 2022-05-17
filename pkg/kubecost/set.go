package kubecost

import "encoding"

type Set interface {
	CloneSet() Set
	IsEmpty() bool
	GetWindow() Window

	// Representations
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}
