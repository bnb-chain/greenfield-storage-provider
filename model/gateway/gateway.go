package gateway

import "encoding/xml"

// ErrorResponse is used in gateway error response
type ErrorResponse struct {
	XMLName xml.Name `xml:"Error"`
	Code    int32    `xml:"Code"`
	Message string   `xml:"Message"`
}
