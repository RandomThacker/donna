package entity

// Health is the domain liveness result.
type Health struct {
	Status string
}

// Readiness is the domain readiness result.
type Readiness struct {
	Status   string
	Database string
}

// Version is the domain build metadata.
type Version struct {
	Version   string
	Commit    string
	BuildTime string
}
