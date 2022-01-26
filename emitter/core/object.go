package core

type Object struct {
	memberTracker memberTracker
}

func NewObject(members []*Member) *Object {
	obj := &Object{memberTracker: newMemberTracker()}
	for _, m := range members {
		obj.AddMember(m)
	}

	return obj
}

func (o *Object) AddMember(m *Member) {
	o.memberTracker.AddMember(m)
}

func (o *Object) AllMembers() map[string]*Member {
	return o.memberTracker.AllMembers()
}

func (o *Object) WaitForMember(name string) *Member {
	return o.memberTracker.WaitForMember(name)
}
