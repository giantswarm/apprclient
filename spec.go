package apprclient

const (
	httpClientTimeout = 5

	// status returned on a successful GET request to a CNR server.
	okStauts = "ok"
	// status returned on a successful DELETE request to a CNR server.
	deletedStatus = "deleted"
)

// Interface describes the methods provided by the appr client.
type Interface interface {
	GetReleaseVersion(name, channel string) (string, error)
	PullChartTarball(name, channel string) (string, error)
}

// Channel represents a CNR channel.
type Channel struct {
	Current string `json:"current"`
}
