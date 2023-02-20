package https

const (
	CodeOk = 20000
)

const (
	CtxResponseKey = "_CTX_RESPONSE"
)

// https://google.github.io/styleguide/jsoncstyleguide.xml#JSON_Structure_&_Reserved_Property_Names
type Response struct {
	Error *Error      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}
