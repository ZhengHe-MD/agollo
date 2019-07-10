package agollo

type ChangeEventObserver interface {
	HandleChangeEvent(event *ChangeEvent)
}
