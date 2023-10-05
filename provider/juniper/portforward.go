package juniper

type PortForward struct {
	Name         string
	Target       string
	Port         int32
	IpAddr       string
	Proto        string
	Policy       string
	ExternalZone string
	InternalZone string
	RuleSet      string
}

func NewPortForward() (PortForward, error) {
	return PortForward{}, nil
}
