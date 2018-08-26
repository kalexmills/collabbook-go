package data

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"time"
)

type Repo struct {
	items  map[uint64]*Item
	boards map[string]map[uint64]bool
}

const DefaultBoard = "My board"
const ArchiveBoard = "archive"

func NewRepo() *Repo {
	result := new(Repo)
	result.items = make(map[uint64]*Item)
	result.boards = make(map[string]map[uint64]bool)

	result.boards[DefaultBoard] = make(map[uint64]bool)
	result.boards[ArchiveBoard] = make(map[uint64]bool)

	return result
}

func (store *Repo) Item(id uint64) *Item {
	result, _ := store.items[id]
	return result
}

func (store *Repo) Boards() []string {
	keys := make([]string, len(store.boards))

	i := 0
	for k := range store.boards {
		keys[i] = k
		i++
	}
	return keys
}

func (store *Repo) IdsInBoard(name string) []uint64 {
	if store.boards[name] == nil {
		return nil
	}
	result := make([]uint64, len(store.boards[name]))

	i := 0
	for id := range store.boards[name] {
		result[i] = id
		i += 1
	}

	return result
}

func (store *Repo) ActiveItems() []*Item {
	archive := store.boards[ArchiveBoard]

	result := make([]*Item, 8)
	for id, item := range store.items {
		if !archive[id] {
			result = append(result, item)
		}
	}
	return result
}

func (store *Repo) ItemsInBoards(boards ...string) []*Item {
	if len(boards) == 0 {
		return nil
	}
	// Count total items in all requested boards
	size := 0
	for _, board := range boards {
		size += len(store.boards[board])
	}
	result := make([]*Item, size)

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

func (store *Repo) ToggleItemIsStarred(id uint64) {
	it, ok := store.items[id]
	if ok {
		it.SetStarred(!it.IsStarred())
	}
}

func (store *Repo) ToggleTaskIsComplete(id uint64) error {
	it, ok := store.items[id]
	if ok {
		if !it.IsTask() {
			return &NotATaskError{id}
		}
		it.SetComplete(!it.IsComplete())
	}
	return nil
}

func (store *Repo) DeleteItem(id uint64) *Item {
	it, _ := store.items[id]
	store.items[id] = nil
	return it
}

func (store *Repo) MakeNote(desc string, boards ...string) *Item {
	return store.makeItem(desc, boards...)
}

func (store *Repo) MakeTask(desc string, boards ...string) *Item {
	result := store.makeItem(desc, boards...)
	result.flags = taskFlag

	return result
}

var nextId uint64 = 0 // Using this here works only since collabbook only has one Repo open per execution

func (store *Repo) makeItem(desc string, boards ...string) *Item {
	result := &Item{nextId, 0, time.Now(), desc}
	nextId += 1

	store.items[result.Id] = result

	if len(boards) == 0 {
		store.AddItemToBoard(result, DefaultBoard)
	} else {
		for _, board := range boards {
			store.AddItemToBoard(result, board)
		}
	}
	return result
}

func (store *Repo) AddItemToBoard(item *Item, boardname string) {
	_, ok := store.boards[boardname]
	if !ok {
		store.boards[boardname] = make(map[uint64]bool)
	}
	store.boards[boardname][item.Id] = true
}

type NotATaskError struct {
	id uint64
}

func (err *NotATaskError) Error() string {
	return fmt.Sprintf("%d is not marked as a task", err.id)
}

//----------------------------------------------------------------------------------------------------------------------
// TextMarshaler related code below
//----------------------------------------------------------------------------------------------------------------------

type CouldNotParse struct{}

func (err *CouldNotParse) Error() string {
	return fmt.Sprintf("Could not unmarshal Repo.")
}

type UnexpectedEndOfInput struct{}

func (err *UnexpectedEndOfInput) Error() string {
	return fmt.Sprintf("Unexpected end of input")
}

func (store *Repo) UnmarshalText(text []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = &CouldNotParse{}
		}
	}()

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

func (store *Repo) MarshalText() (text []byte, err error) {
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
		_, err = buf.WriteString(strconv.FormatUint(itemid, 10))
		_, err = buf.WriteRune('\n')
	}
	_, err = buf.WriteString("---\n")
	return
}

func unmarshalBoard(s *bufio.Scanner, store *Repo) (done bool, err error) {
	if !s.Scan() {
		return true, nil
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

func marshalItem(it *Item, buf *bytes.Buffer) (err error) {
	_, err = buf.WriteString(strconv.FormatUint(it.Id, 10))
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

	// Write Item separator
	_, err = buf.WriteString("---\n")
	return
}

func max(a, b uint64) uint64 {
	if (a > b) {
		return a
	}
	return b
}

func unmarshalItem(s *bufio.Scanner, store *Repo) (done bool, err error) {
	it := new(Item)
	s.Scan()
	tok := s.Text()
	if tok == "=====" {
		return true, nil
	}

	it.Id, err = strconv.ParseUint(tok, 10, 64)
	store.items[it.Id] = it
	nextId = max(it.Id + 1, nextId)

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
