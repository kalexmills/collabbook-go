package app

import (
	"time"
	"fmt"
	"bytes"
	"bufio"
	"strconv"
)

type notATaskError struct {
	id uint64
}

func (err *notATaskError) Error() string {
	return fmt.Sprintf("%d is not marked as a task", err.id)
}

type UnexpectedEndOfInput struct {}

func (err *UnexpectedEndOfInput) Error() string {
	return fmt.Sprintf("Unexpected end of input")
}

const (
	taskFlag = 1 << iota
	starFlag
	completeFlag
)

type item struct {
	Id uint64
	flags byte
	CreatedUTC time.Time
	Desc string
}

func (it *item) IsTask() bool {
	return it.flags & taskFlag > 0
}

func (it *item) IsStarred() bool {
	return it.flags & starFlag > 0
}

func (it *item) IsComplete() bool {
	return (it.flags & (completeFlag | taskFlag)) > 0
}

func (it *item) SetStarred(value bool) {
	setFlag(&it.flags, starFlag, value)
}

func (it *item) SetComplete(value bool) {
	setFlag(&it.flags, completeFlag, value)
}

func setFlag(flag *byte, mask byte, value bool) {
	var x byte = 0
	if value { x = 1 }
	*flag ^= (-x ^ *flag) & mask  // set the bit via mask
}

type itemstore struct {
	items map[uint64]*item
	boards map[string]map[uint64]bool
}

const defaultBoard = "DEFAULT_BOARD"
const archiveBoard = "ARCHIVE_BOARD"

func New() *itemstore {
	result := new(itemstore)
	result.items = make(map[uint64]*item)
	result.boards = make(map[string]map[uint64]bool)

	result.boards[defaultBoard] = make(map[uint64]bool)
	result.boards[archiveBoard] = make(map[uint64]bool)

	return result
}

func (store *itemstore) GetItem(id uint64) *item {
	result, _ := store.items[id]
	return result
}

func (store *itemstore) ActiveItems() []*item {
	archive := store.boards[archiveBoard]

	result := make([]*item, 8)
	for id, item := range store.items {
		if !archive[id] {
			result = append(result, item)
		}
	}
	return result
}

func (store *itemstore) ItemsInBoards(boards ...string) []*item {
	if len(boards) == 0 { return nil }
	// Count total items in all requested boards
	size := 0
	for _, board := range boards {
		size += len(store.boards[board])
	}
	result := make([]*item, size)

	// Add all items to result slice
	i := 0
	for _, board := range boards {
		items, ok := store.boards[board]
		if ok {
			for itemid := range items {
				result[i] = store.items[itemid]
				i += 1
			}
		}
	}
	return result
}

func (store *itemstore) ToggleItemIsStarred(id uint64) {
	it, ok := store.items[id]
	if ok {
		it.SetStarred(!it.IsStarred())
	}
}

func (store *itemstore) ToggleTaskIsComplete(id uint64) error {
	it, ok := store.items[id]
	if ok {
		if !it.IsTask() { return &notATaskError{id} }
		it.SetComplete(!it.IsComplete())
	}
	return nil
}

func (store *itemstore) DeleteItem(id uint64) *item {
	it, _ := store.items[id]
	store.items[id] = nil
	return it
}

func (store *itemstore) MakeTask(desc string) *item {
	result := store.MakeItem(desc)
	result.flags = taskFlag

	return result
}

var nextId uint64 = 0   // Using this here works only since collabbook only has one itemstore open per execution

func (store *itemstore) MakeItem(desc string, boards ...string) *item {
	result := &item{nextId, 0, time.Now(), desc}
	nextId += 1

	store.items[result.Id] = result

	if len(boards) == 0 {
		store.AddItemToBoard(result, defaultBoard)
	} else {
		for _, board := range boards {
			store.AddItemToBoard(result, board)
		}
	}
	return result
}

func (store *itemstore) AddItemToBoard(item *item, boardname string) {
	_, ok := store.boards[boardname]
	if !ok {
		store.boards[boardname] = make(map[uint64]bool)
	}
	store.boards[boardname][item.Id] = true
}

func (store *itemstore) UnmarshalText(text []byte) (err error) {
	s := bufio.NewScanner(bytes.NewReader(text))

	done := false
	for !done {
		done, err = unmarshalItem(s, store)
	}
	done = false
	for !done {
		done, err = unmarshalBoard(s, store)
	}
	return
}

func (store *itemstore) MarshalText() (text []byte, err error) {
	buf := new(bytes.Buffer)

	for _, item := range store.items {
		err = marshalItem(item, buf)
	}
	buf.WriteString("=====\n")

	for name, board := range store.boards {
		err = marshalBoard(name, board, buf)
	}

	return buf.Bytes(), err
}

func marshalBoard(name string, board map[uint64]bool, buf *bytes.Buffer) (err error) {
	_, err = buf.WriteString(name)
	_, err = buf.WriteRune('\n')

	for itemid := range board {
		_, err = buf.WriteString(string(itemid))
		_, err = buf.WriteRune('\n')
	}
	_, err = buf.WriteString("---")
	return
}

func unmarshalBoard(s *bufio.Scanner, store *itemstore) (done bool, err error) {
	if !s.Scan() {
		return done, nil
	}
	tok := s.Text()
	board := make(map[uint64]bool, 4)
	store.boards[tok] = board

	for s.Scan() {
		tok = s.Text()
		if tok == "---" {
			return false, err
		}

		var itemid uint64
		itemid, err = strconv.ParseUint(tok, 10, 64)

		board[itemid] = true
	}
	return true, err
}

func marshalItem(it *item, buf *bytes.Buffer) (err error) {
	_, err = buf.WriteString(string(it.Id))
	_, err = buf.WriteRune('\n')

	if it.IsTask() {
		_, err = buf.WriteString("T\n")
		if it.IsComplete() {
			_, err = buf.WriteString("T\n")
		} else {
			_, err = buf.WriteString("F\n")
		}
	} else {
		_, err = buf.WriteString("N\n")
	}

	if it.IsStarred() {
		_, err = buf.WriteString("T\n")
	} else {
		_, err = buf.WriteString("F\n")
	}

	date, err := it.CreatedUTC.MarshalText()

	_, err = buf.Write(date)
	_, err = buf.WriteRune('\n')

	_, err = buf.WriteString(it.Desc)
	_, err = buf.WriteRune('\n')

	// Write item separator
	_, err = buf.WriteString("---\n")
	return
}


func unmarshalItem(s *bufio.Scanner, store *itemstore) (done bool, err error) {
	it := new(item)
	s.Scan()
	tok := s.Text()
	if tok == "=====" {
		return true, nil
	}

	it.Id, err = strconv.ParseUint(tok, 10, 64)
	s.Scan()
	tok = s.Text()
	if tok[0] == 'T' {
		it.flags = taskFlag
		s.Scan()
		tok = s.Text()
		if tok[0] == 'T' {
			it.SetComplete(true)
		}
	}
	s.Scan()
	tok = s.Text()
	if tok[0] == 'T' {
		it.SetStarred(true)
	}

	s.Scan()
	tok = s.Text()
	it.CreatedUTC, err = time.Parse(time.RFC3339, tok)

	s.Scan()
	it.Desc = s.Text()

	// Check separator to see if there are more items to read
	if s.Scan() == false {
		return true, &UnexpectedEndOfInput{}
	}

	err = s.Err()
	return
}
