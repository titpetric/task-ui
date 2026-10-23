package beszel

import (
	"time"
)

// ContainerSample is a record of a host's containers over a span of the
// hub's longer records: when the hub wrote it, and each container's
// averages of the span.
type ContainerSample struct {
	At         time.Time
	Containers []Container
}
