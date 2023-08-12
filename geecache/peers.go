package geecache

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
} //根据传入的key选择相应的节点PeerGetter(即HTTP客户端)

type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
} //从对应的group查找对应的缓存值
