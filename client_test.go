package agollo

import "testing"

type simpleMockObserver struct {}

func (m *simpleMockObserver) HandleChangeEvent(ce *ChangeEvent) {
	// noop
}

func TestObserver(t *testing.T) {
	c := NewClient(&Conf{})

	t.Run("register&recall", func(t *testing.T) {
		var ob = &simpleMockObserver{}

		c.registerObserver(ob)

		if !(len(c.observers) == 1 && c.observers[0] == ob) {
			t.Errorf("observer should be added to client")
		}

		c.recallObserver(ob)

		if len(c.observers) > 0 {
			t.Errorf("observer should be recalled")
		}
	})

	t.Run("getObservers", func(t *testing.T) {
		num := 10
		for i := 0; i < num; i++ {
			ob := &simpleMockObserver{}
			c.registerObserver(ob)
		}

		obs := c.getObservers()

		if len(obs) != num {
			t.Errorf("should have %d observers but got %d", num, len(obs))
		}
	})
}
