package geecache

type ByteView struct {
	b []byte //储存一个缓存值
}

func (v ByteView) Len() int { //实现LRU中的接口
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b) //返回拷贝值防止从外部更改
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
