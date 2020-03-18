package main

//go:generate gml -type=example -file-name=u_can_set_file_name_or_by_default__-type_gml.go
type example byte

const (
	ErrCode200 example = 0 // request ok
	ErrCode400 example = 1 // request not found
	ErrCode500 example = 2 // request failed
)
