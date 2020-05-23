package skarner

type Scanner interface {
	Scan(model interface{}) error
}
