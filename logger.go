package agollo

type AgolloLogger interface {
	Printf(format string, v ...interface{})
}
