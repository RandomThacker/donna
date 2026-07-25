package model

import "github.com/RandomThacker/donna/services/api/internal/entity"

// HealthResponse is the HTTP DTO for liveness.
type HealthResponse struct {
	Status string `json:"status"`
}

// ReadyResponse is the HTTP DTO for readiness.
type ReadyResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

// VersionResponse is the HTTP DTO for build metadata.
type VersionResponse struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

// HealthFromEntity maps a domain health entity to a transport model.
func HealthFromEntity(e entity.Health) HealthResponse {
	return HealthResponse{Status: e.Status}
}

// ReadyFromEntity maps a domain readiness entity to a transport model.
func ReadyFromEntity(e entity.Readiness) ReadyResponse {
	return ReadyResponse{Status: e.Status, Database: e.Database}
}

// VersionFromEntity maps a domain version entity to a transport model.
func VersionFromEntity(e entity.Version) VersionResponse {
	return VersionResponse{
		Version:   e.Version,
		Commit:    e.Commit,
		BuildTime: e.BuildTime,
	}
}
