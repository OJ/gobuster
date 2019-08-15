package gobusters3

import "encoding/xml"

type AWSError struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	RequestID string   `xml:"RequestId"`
	HostID    string   `xml:"HostId"`
}

// parse only a subset of returned properties
type AWSListing struct {
	XMLName     xml.Name `xml:"ListBucketResult"`
	Name        string   `xml:"Name"`
	IsTruncated string   `xml:"IsTruncated"`
	Contents    []struct {
		Key          string `xml:"Key"`
		LastModified string `xml:"LastModified"`
		Size         int    `xml:"Size"`
	} `xml:"Contents"`
}
