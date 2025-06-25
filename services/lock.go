package services

import (
	"errors"
	"sync"

	"github.com/ahmetb/go-linq/v3"
)

var DefaultLibraryLockPool = LibraryLock{LockLibraryIds: []uint{}}
var LibraryLockError = errors.New("library is locked")

type LibraryLock struct {
	LockLibraryIds []uint
	sync.Mutex
}

func (l *LibraryLock) TryToLock(libraryId uint) bool {
	l.Lock()
	defer l.Unlock()
	for _, id := range l.LockLibraryIds {
		if id == libraryId {
			return false
		}
	}
	l.LockLibraryIds = append(l.LockLibraryIds, libraryId)
	return true
}

func (l *LibraryLock) TryToUnlock(libraryId uint) {
	l.Lock()
	defer l.Unlock()
	linq.From(l.LockLibraryIds).Where(func(i interface{}) bool {
		return i.(uint) != libraryId
	}).ToSlice(&l.LockLibraryIds)
	return
}
