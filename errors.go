package main

import "errors"

var (
	ErrRequestFailed = errors.New("server returned an error status code")
	ErrParsingFailed = errors.New("unable to parse data")
)
