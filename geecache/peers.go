package geecache

import pb "geecache/geecachepb"

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) //用于根据传入的key选择响应节点PeerGetter
}

// PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error //用于从对应group查找缓存值，PeerGetter对应于上述流程中的HTTP客户端
}
