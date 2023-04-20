package p2p

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDNSParser(t *testing.T) {
	dns := "localhost"
	hosts, err := net.LookupHost(dns)
	require.NoError(t, err)
	for _, host := range hosts {
		fmt.Println(host)
	}
}

func TestBootstrapParser(t *testing.T) {
	testCases := []struct {
		name      string
		bootstrap []string
	}{
		{
			name: "domain",
			bootstrap: []string{
				"16Uiu2HAmBzdPttaxicSDEf5Kq1XBnoH97wRFA8aiWEnYc2hp2ZHW@localhost:9933",
			},
		},
		{
			name: "ip",
			bootstrap: []string{
				"16Uiu2HAmBzdPttaxicSDEf5Kq1XBnoH97wRFA8aiWEnYc2hp2ZHW@0.0.0.0:9933",
			},
		},
		{
			name: "domain and ip mix",
			bootstrap: []string{
				"16Uiu2HAmBzdPttaxicSDEf5Kq1XBnoH97wRFA8aiWEnYc2hp2ZHW@localhost:9933",
				"16Uiu2HAmBzdPttaxicSDEf5Kq1XBnoH97wRFA8aiWEnYc2hp2ZHW@0.0.0.0:9933",
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			peersIDs, addrs, err := MakeBootstrapMultiaddr(testCase.bootstrap)
			require.NoError(t, err)
			for i, addr := range addrs {
				fmt.Println("peerID: " + peersIDs[i].String() + ", addr" + addr.String())
			}
		})
	}
}
