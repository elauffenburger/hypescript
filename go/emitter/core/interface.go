package core

type Interface struct {
	memberTracker memberTracker

	Name string
}

func NewInterface(name string, members []*Member) *Interface {
	i := &Interface{Name: name, memberTracker: newMemberTracker()}
	for _, m := range members {
		i.AddMember(m)
	}

	return i
}

func (i *Interface) AddMember(m *Member) {
	i.memberTracker.AddMember(m)
}

func (i *Interface) AllMembers() map[string]*Member {
	return i.memberTracker.AllMembers()
}

func (i *Interface) WaitForMember(name string) *Member {
	return i.memberTracker.WaitForMember(name)
}
