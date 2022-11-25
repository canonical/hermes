package perf

type GroupReadContent struct {
	TimeEnabled uint64
	TimeRunning uint64
	Values      []struct {
		Value uint64
		ID    uint64
	}
}

type Group struct {
	ReadFormat  ReadFormat
	Options     Options
	ClockID     int32
	Attrs       []*Attr
	needRingBuf bool
}

func (group *Group) hasAttrs() bool {
	return len(group.Attrs) > 0
}

func (group *Group) AddAttrs(configurators ...AttrConfigurator) {
	for _, configurator := range configurators {
		attr := new(Attr)
		attr.ReadFormat = group.ReadFormat
		attr.Options = group.Options
		attr.ClockID = group.ClockID
		configurator.Configure(attr)

		if attr.Sample != 0 {
			group.needRingBuf = true
		}
		if !group.hasAttrs() {
			attr.ReadFormat.Group = true
		}
		group.Attrs = append(group.Attrs, attr)
	}
}

func (group *Group) GetLeaderAttr() *Attr {
	if !group.hasAttrs() {
		return nil
	}
	return group.Attrs[0]
}

func (group *Group) GetFollowerAttrs() []*Attr {
	return group.Attrs[1:]
}
