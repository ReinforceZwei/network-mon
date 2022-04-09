package pingwrap

type PingWrap interface {
	PingOnce(string) bool
}
