package storage

import "fmt"

// Gauge defines gauge type.
type Gauge float64

func (g Gauge) String() string {
	return fmt.Sprintf("%f", g)
}

// Counter defines counter type.
type Counter int64

func (c Counter) String() string {
	return fmt.Sprintf("%d", c)
}

// Types declares possible types of storage
type Types interface {
	Gauge | Counter
}
