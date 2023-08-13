package geecache

import (
	pb "geecache/geecachepb"
)

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
} //根据传入的key选择相应的节点PeerGetter(即HTTP客户端)

type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
} //从对应的group查找对应的缓存值
