package geecache

import (
	"fmt"
	consistenthash "geecache/consistencehash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

type HTTPPool struct { //承载节点间HTTP通信的核心数据结构
	self        string //记录自己的地址，包括主机名/IP 和端口
	basePath    string //节点间通讯地址的前缀
	mu          sync.Mutex
	peers       *consistenthash.Map    //根据具体的key选择节点
	httpGetters map[string]*httpGetter //映射远程节点httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Set(peers ...string) { //实例化一致性哈希算法
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)                                    //向哈希环中添加节点
	p.httpGetters = make(map[string]*httpGetter, len(peers)) //初始化一个存储HTTP客户端的map

	for _, peer := range peers { //为每个节点创建HTTP客户端httpGetter
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) { //包装一致性哈希的Get()方法，根据具体的key选择节点，返回节点对应的HTTP客户端
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) { //判断访问路径的前缀是否是 basePath
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	//访问路径格式为 /<basepath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2) //切割访问路径后两部分
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)

	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice()) //将缓存值作为 httpResponse 的 body 返回
}

type httpGetter struct {
	baseURL string //将要访问的远程节点的地址
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) { //使用 http.Get() 方式获取返回值，将返回值转换为[]bytes类型
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key)) //拼接成完整的url
	res, err := http.Get(u)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK { //检查响应状态码是否ok
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	//读取响应体并将其转化为[]byte类型
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)
