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
	}
	if rpcx.TransferType == BIDIRECTIONAL { //双向的
		ch := make(chan *protocol.Message)
		rpcxClient = client.NewBidirectionalXClient(rpcx.Module, client.Failtry, client.RandomSelect, d, client.DefaultOption, ch)
		go func() {
			//处理游戏服的推送
			HandleGamePush := func(userid, msg string) {
				client := GetClientManager().GetClient(userid)
				client.Send_log([]byte(msg))
			}
			for msg := range ch {
				content := string(msg.Payload)
				arr := strings.Split(content, ":")
				userid, msg := arr[0], arr[1]
				HandleGamePush(userid, msg)
			}
		}()
	} else { //单向的
		rpcxClient = client.NewXClient(rpcx.Module, client.Failtry, client.RandomSelect, d, client.DefaultOption)
	}
	return rpcxClient
}
