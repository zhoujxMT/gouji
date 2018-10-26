package frame

import (
	"strings"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
)

type RpcxClient struct {
	DiscoveryType int
	Rpcx
	Module       string
	TransferType int
}

func NewRpcxClient(rpcx *RpcxClient) client.XClient {
	var rpcxClient client.XClient
	var d client.ServiceDiscovery
	if rpcx.DiscoveryType == Discovery_Peer2Peer {
		d = client.NewPeer2PeerDiscovery("tcp@"+*rpcx.Addr, "")
	} else if rpcx.DiscoveryType == Discovery_Etcd {
		d = client.NewEtcdDiscovery(*rpcx.BasePath, rpcx.Module, []string{*rpcx.EtcdAddr}, nil)
	} else if rpcx.DiscoveryType == Discovery_MultipleServers {
		serversAddr := strings.Split(*rpcx.Addr, "|")
		kvPairs := []*client.KVPair{}
		for _, serverAddr := range serversAddr {
			kvPairs = append(kvPairs, &client.KVPair{Key: serverAddr})
		}
		d = client.NewMultipleServersDiscovery(kvPairs)
	}
	if rpcx.TransferType == BIDIRECTIONAL { //双向的
		ch := make(chan *protocol.Message)
		rpcxClient = client.NewBidirectionalXClient(rpcx.Module, client.Failtry, client.RandomSelect, d, client.DefaultOption, ch)
		go func() {
			//处理游戏服的推送
			HandleGamePush := func(userid, funcName, content string) {
				cl := GetClientManager().GetClient(userid)
				if cl != nil {
					msg := cl.Format(funcName, content)
					cl.Log_Push(msg)
				}
			}
			for msg := range ch {
				content := string(msg.Payload)
				arr := strings.Split(content, "&")
				userid, funcName, content := arr[0], arr[1], arr[2]
				HandleGamePush(userid, funcName, content)
			}
		}()
	} else { //单向的
		if rpcx.DiscoveryType == Discovery_MultipleServers {
			rpcxClient = client.NewXClient(rpcx.Module, client.Failover, client.RoundRobin, d, client.DefaultOption)
		} else {
			rpcxClient = client.NewXClient(rpcx.Module, client.Failtry, client.RandomSelect, d, client.DefaultOption)
		}
	}
	return rpcxClient
}
