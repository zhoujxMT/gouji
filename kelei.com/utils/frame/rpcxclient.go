package frame

import (
	"context"
	"strings"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"kelei.com/utils/logger"
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
		rpcxClient.SetSelector(&alwaysFirstSelector{})
		go func() {
			defer func() {
				if p := recover(); p != nil {
					logger.Errorf("[recovery] HandleGamePush : %v", p)
				}
			}()
			//处理游戏服的推送
			HandleGamePush := func(userid, funcName, content string) {
				cl := GetClientManager().GetClient(userid)
				if cl != nil {
					msg := cl.Format(funcName, content)
					if funcName == "MatchEnd" {
						cl.setCurrentRoomID("-1")
					} else {
						cl.Log_Push(msg)
					}
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
			rpcxClient = client.NewXClient(rpcx.Module, client.Failtry, client.RoundRobin, d, client.DefaultOption)
		} else {
			rpcxClient = client.NewXClient(rpcx.Module, client.Failtry, client.RandomSelect, d, client.DefaultOption)
		}
	}
	d.Close()
	return rpcxClient
}

type alwaysFirstSelector struct {
	servers []string
}

func (s *alwaysFirstSelector) Select(ctx context.Context, servicePath, serviceMethod string, args interface{}) string {
	var ss = s.servers
	if len(ss) == 0 {
		return ""
	}
	return ss[0]
}

func (s *alwaysFirstSelector) UpdateServer(servers map[string]string) {
	var ss = make([]string, 0, len(servers))
	for k := range servers {
		ss = append(ss, k)
	}
	//	sort.Slice(ss, func(i, j int) bool {
	//		return strings.Compare(ss[i], ss[j]) <= 0
	//	})
	s.servers = ss
}
