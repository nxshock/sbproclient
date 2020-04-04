package sbproclient

var (
	messagePrefix []byte = []byte("<ms>")
	messageSuffix []byte = []byte("</me>")
)

var (
	requestSymbols []byte = []byte("newsymbols^-^-")
)

var (
	begin = []byte("<st>")
	end   = []byte("</st>")
)
