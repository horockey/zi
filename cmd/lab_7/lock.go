package main

type Lock struct {
	User      string `json:"user"`
	UntilUnix int64  `json:"until"`
}
