package main

//go:generate go-enum

// ENUM(public_data = 1, secret, top_secret)
type AccessLevel uint8
