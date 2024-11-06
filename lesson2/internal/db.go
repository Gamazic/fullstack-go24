package internal

type Lesson struct {
	Id       int `gorm:"column:lesson_id"`
	Name     string
	Students string
}

type LessonDB interface {
	GetLessons(limit, offset int) ([]Lesson, error)
	GetLessonById(id int) (Lesson, bool, error)
	PutLesson(lesson Lesson) error
	DeleteLesson(id int) error
}
