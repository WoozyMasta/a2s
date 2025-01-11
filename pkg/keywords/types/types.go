/*
Package types implements type representation for various data encoded in responses,
such as server language, OS, game type, or server state.

MarshalJSON enabled by default for types, for disable it use build tag a2s_no_marshal

	go build -tags=a2s_no_marshal
*/
package types
