package inmemory

import (
	"lessonDb/internal"
	"sync"
)

type InmemoryDb struct {
	lessons map[int]internal.Lesson
	idToIdx map[int]int
	idx     []idx
	mu      sync.RWMutex
}

type idx struct {
	Id        int
	IsDeleted bool
}

func NewInmemoryDb() *InmemoryDb {
	return &InmemoryDb{
		lessons: make(map[int]internal.Lesson),
		idx:     make([]idx, 0),
		idToIdx: make(map[int]int),
	}
}

func (i *InmemoryDb) GetLessons(limit, offset int) ([]internal.Lesson, error) {
	start := min(offset, len(i.idx))
	end := min(offset+limit, len(i.idx))
	lessonsIds := i.idx[start:end]
	lessons := []internal.Lesson{}
	for _, indx := range lessonsIds {
		if indx.IsDeleted {
			continue
		}
		lessons = append(lessons, i.lessons[indx.Id])
	}
	return lessons, nil
}

func (i *InmemoryDb) GetLessonById(id int) (internal.Lesson, bool, error) {
	l, ok := i.lessons[id]
	if !ok {
		return internal.Lesson{}, false, nil
	}
	return l, true, nil
}

func (i *InmemoryDb) PutLesson(lesson internal.Lesson) error {
	if _, ok := i.lessons[lesson.Id]; ok {
		i.lessons[lesson.Id] = lesson
		return nil
	}
	i.lessons[lesson.Id] = lesson
	i.idx = append(i.idx, idx{Id: lesson.Id, IsDeleted: false})
	i.idToIdx[lesson.Id] = len(i.idx) - 1
	return nil
}

func (i *InmemoryDb) DeleteLesson(id int) error {
	if _, ok := i.lessons[id]; !ok {
		return nil
	}
	delete(i.lessons, id)
	indx := i.idToIdx[id]
	i.idx[indx].IsDeleted = true
	return nil
}
