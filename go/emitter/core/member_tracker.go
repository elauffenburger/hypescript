package core

import "sync"

type memberTracker struct {
	lock      sync.Mutex
	listeners map[string][]chan *Member

	members map[string]*Member
}

func newMemberTracker() memberTracker {
	return memberTracker{
		listeners: map[string][]chan *Member{},
		members:   map[string]*Member{},
	}
}

func (t *memberTracker) AddMember(m *Member) {
	t.lock.Lock()

	t.members[*m.Name()] = m
	t.resolveMember(m)

	t.lock.Unlock()
}

func (t *memberTracker) AllMembers() map[string]*Member {
	return t.members
}

func (t *memberTracker) WaitForMember(name string) *Member {
	if m, ok := t.members[name]; ok {
		return m
	}

	c := make(chan *Member)

	t.lock.Lock()

	if _, ok := t.listeners[name]; !ok {
		t.listeners[name] = []chan *Member{}
	}

	t.listeners[name] = append(t.listeners[name], c)

	t.lock.Unlock()

	return <-c
}

func (t *memberTracker) resolveMember(m *Member) {
	listeners := t.listeners[*m.Name()]
	if listeners == nil {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(listeners))

	for _, l := range listeners {
		l := l

		go func() {
			l <- m
			close(l)

			wg.Done()
		}()
	}

	wg.Wait()
}

type Member struct {
	Field *ObjectTypeField
}

func (m *Member) Name() *string {
	if m.Field != nil {
		return &m.Field.Name
	}

	panic("unknown member type")
}

func (m *Member) Type() *TypeSpec {
	if m.Field != nil {
		return m.Field.Type
	}

	panic("unknown member type")
}

type HasMembers interface {
	AddMember(m *Member)
	AllMembers() map[string]*Member
	WaitForMember(name string) *Member
}
