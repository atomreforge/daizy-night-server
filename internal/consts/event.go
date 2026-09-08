package consts

type EventType string

const (
	MPI  EventType = "major-public-incident" // negative
	MPE  EventType = "major-public-event"    // positive or neutral
	MPHE EventType = "major-public-health-emergency"
	MPSE EventType = "major-public-security-emergency"
)
