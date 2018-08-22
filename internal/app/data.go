package app

import (
	"time"
	"fmt"
)

type notATaskError struct {
	id int64
}

func (err *notATaskError) Error() string {
	return fmt.Sprintf("%d is not marked as a task", err.id)
}

type item struct {
	Id int64
	Desc string
	UTCDate time.Time
	IsStarred bool
	IsTask bool
	IsCompleted bool
}

type itemstore struct {
	items map[int64]*item
	boards map[string]map[int64]bool
}

const defaultBoard = "DEFAULT_BOARD"
const archiveBoard = "ARCHIVE_BOARD"

func NewItemstore() *itemstore {
	result := new(itemstore)
	result.items = make(map[int64]*item)
	result.boards = make(map[string]map[int64]bool)

	result.boards[defaultBoard] = make(map[int64]bool)
	result.boards[archiveBoard] = make(map[int64]bool)

	return result
}

func (store *itemstore) GetItem(id int64) *item {
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

func (store *itemstore) ToggleItemIsStarred(id int64) {
	it, ok := store.items[id]
	if ok {
		it.IsStarred = !it.IsStarred
	}
}

func (store *itemstore) ToggleTaskIsComplete(id int64) error {
	it, ok := store.items[id]
	if ok {
		if !it.IsTask { return &notATaskError{id} }
		it.IsCompleted = !it.IsCompleted
	}
	return nil
}

func (store *itemstore) DeleteItem(id int64) *item {
	it, _ := store.items[id]
	store.items[id] = nil
	return it
}

func (store *itemstore) MakeTask(desc string) *item {
	result := store.MakeItem(desc)
	result.IsTask = true
	return result
}

func (store *itemstore) MakeItem(desc string, boards ...string) *item {
	result := new(item)
	result.Id = getNextId()
	result.Desc = desc
	result.UTCDate = time.Now()

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
	board := store.lazyGetBoard(boardname)
	board[item.Id] = true
}

func (store *itemstore) lazyGetBoard(boardname string) map[int64]bool {
	_, ok := store.boards[boardname]
	if !ok {
		store.boards[boardname] = make(map[int64]bool)
	}
	return store.boards[boardname]
}

var nextId int64 = 0   // This pattern works only since collabbook only has one itemstore open per execution
func getNextId() int64 {
	result := nextId
	nextId += 1
	return result
}

