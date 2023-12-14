package types

import (
	"encoding/base64"
	"encoding/xml"
)

func (m GfSpListObjectsByBucketNameResponse) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type Alias GfSpListObjectsByBucketNameResponse
	// Create a new struct with Base64-encoded Checksums field
	responseAlias := Alias(m)
	for _, o := range responseAlias.Objects {
		for i, c := range o.ObjectInfo.Checksums {
			o.ObjectInfo.Checksums[i] = []byte(base64.StdEncoding.EncodeToString(c))
		}
	}
	return e.EncodeElement(responseAlias, start)
}

func (m GfSpGetObjectMetaResponse) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type Alias GfSpGetObjectMetaResponse
	// Create a new struct with Base64-encoded Checksums field
	responseAlias := Alias(m)
	o := responseAlias.Object
	if o != nil && o.ObjectInfo != nil && o.ObjectInfo.Checksums != nil {
		for i, c := range o.ObjectInfo.Checksums {
			o.ObjectInfo.Checksums[i] = []byte(base64.StdEncoding.EncodeToString(c))
		}
	}
	return e.EncodeElement(responseAlias, start)
}

type GroupEntry struct {
	Id    uint64
	Value *Group
}

func (m GfSpListGroupsByIDsResponse) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(m.Groups) == 0 {
		return nil
	}

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for k, v := range m.Groups {
		e.Encode(GroupEntry{Id: k, Value: v})
	}

	return e.EncodeToken(start.End())
}

type ObjectEntry struct {
	Id    uint64
	Value *Object
}

func (m GfSpListObjectsByIDsResponse) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(m.Objects) == 0 {
		return nil
	}

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for k, o := range m.Objects {
		if o != nil && o.ObjectInfo != nil && o.ObjectInfo.Checksums != nil {
			for i, c := range o.ObjectInfo.Checksums {
				o.ObjectInfo.Checksums[i] = []byte(base64.StdEncoding.EncodeToString(c))
			}
		}
		e.Encode(ObjectEntry{Id: k, Value: o})
	}

	return e.EncodeToken(start.End())
}

type BucketEntry struct {
	Id    uint64
	Value *Bucket
}

func (m GfSpListBucketsByIDsResponse) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(m.Buckets) == 0 {
		return nil
	}

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for k, v := range m.Buckets {
		e.Encode(BucketEntry{Id: k, Value: v})
	}

	return e.EncodeToken(start.End())
}
