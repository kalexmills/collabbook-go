package data

import "time"

type Item struct {
	Id         uint64
	flags      byte
	CreatedUTC time.Time
	Desc       string
}

const (
	taskFlag = 1 << iota
	starFlag
	completeFlag
)

func setFlag(flag *byte, mask byte, value bool) {
	var x byte = 0
	if value {
		x = 1
	}
	*flag ^= (-x ^ *flag) & mask // set the bit via mask
}

func (it *Item) IsTask() bool {
	return it.flags&taskFlag > 0
}

func (it *Item) IsStarred() bool {
	return it.flags&starFlag > 0
}

func (it *Item) IsComplete() bool {
	return it.IsTask() && (it.flags & (completeFlag)) > 0
}

func (it *Item) SetStarred(value bool) {
	setFlag(&it.flags, starFlag, value)
}

func (it *Item) SetComplete(value bool) {
	setFlag(&it.flags, completeFlag, value)
}
