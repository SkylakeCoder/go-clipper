package clipper

type msgType byte

const (
	MSG_NULL msgType = iota
	MSG_REGISTER
	MSG_SET_CLIPPER_INFO
	MSG_GET_CLIPPER_INFO
)

type commonReq struct {
	msgID msgType
	msgLen uint32
}

type respGetClipperInfo struct {
	addr string
	path string
}
