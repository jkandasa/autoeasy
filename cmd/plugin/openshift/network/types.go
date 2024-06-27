package network

import (
	"net"
	"time"
)

type DiagnoseNodeData struct {
	Index                int
	NodeName             string                    `json:"nodeName"`
	NodeNameSimple       string                    `json:"nodeNameSimple"`
	PodName              string                    `json:"podName"`
	PortForwardCloseFunc func()                    `json:"-"`
	PodIP                string                    `json:"podIP"`
	PodHttpUrl           string                    `json:"podHttpUrl"`
	LocalHostPort        uint32                    `json:"localHostPort"`
	HostIP               string                    `json:"hostIP"`
	DiagnoseResponse     []NetworkDiagnoseResponse `json:"diagnoseResponse"`
	Roles                []string                  `json:"roles"`
}

type NetworkDiagnoseResponse struct {
	IPerf3 *CmdResponse `json:"iperf3,omitempty"`
	Ping   *CmdResponse `json:"ping,omitempty"`
}

type CmdResponse struct {
	Hostname     string   `json:"hostname"`
	Response     string   `json:"response"`
	ErrorMessage string   `json:"error"`
	IsSuccess    bool     `json:"isSuccess"`
	Options      []string `json:"options"`
}

// ping statistics
type PingStatistics struct {
	PacketsRecv           int
	PacketsSent           int
	PacketsRecvDuplicates int
	PacketLoss            float64
	IPAddr                *net.IPAddr
	Addr                  string
	Rtts                  []time.Duration
	MinRtt                time.Duration
	MaxRtt                time.Duration
	AvgRtt                time.Duration
	StdDevRtt             time.Duration
}

type IPerf3Statistics struct {
	End IPerf3End `json:"end"`
}

type IPerf3End struct {
	Streams []IPerf3Streams `json:"streams"`
}

type IPerf3Streams struct {
	Sender   IPerf3Stream `json:"sender"`
	Receiver IPerf3Stream `json:"receiver"`
}

type IPerf3Stream struct {
	BitsPerSecond float64 `json:"bits_per_second"`
	Bytes         uint64  `json:"bytes"`
	Retransmits   uint64  `json:"retransmits"`
	Seconds       float64 `json:"seconds"`
	Sender        bool    `json:"sender"`
	Socket        uint64  `json:"socket"`
	Start         float64 `json:"start"`
	End           float64 `json:"end"`
}
