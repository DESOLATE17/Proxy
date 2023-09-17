package proxy

type Proxy interface {
	ListenAndServe() error
}
