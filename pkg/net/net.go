package net

// Status represents the different codes that can return from the handlers
type Status int

const (
	// OK code
	OK Status = iota
	// Created code
	Created
	// BadRequest err code
	BadRequest
	// NotFound err code
	NotFound
	// ServerError code
	ServerError
)

// Method represents the different methods that the server can handle
type Method int

const (
	Select Method = iota
	Insert
	Delete
)

// Query represents an encoding type for the tcp handler
type Query struct {
	Method Method
	Key    string
	Value  []byte
}

// Result represents the final result of the tcp handler
type Result struct {
	Status   Status
	Value    []byte
	Duration string
}
