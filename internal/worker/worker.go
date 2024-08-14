package worker

type Worker struct {
	HostId    string
	HostAddr  string
	Topic     string
	Partition string
}

type Key struct {
	Value     string
	Topic     string
	Partition string
}
