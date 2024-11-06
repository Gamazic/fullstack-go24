package gorm

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"lessonDb/internal"
)

type GormDb struct {
	db *gorm.DB
}

func NewGormDb() (*GormDb, error) {
	db, err := gorm.Open(mysql.Open("root:pass@tcp(localhost:3306)/mydb"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &GormDb{db: db}, nil
}

func (g *GormDb) GetLessons(limit, offset int) ([]internal.Lesson, error) {
	var lessons []internal.Lesson
	g.db.Find(&lessons).Limit(limit).Offset(offset)
	return lessons, nil
}

func (g *GormDb) GetLessonById(id int) (internal.Lesson, bool, error) {
	var lesson internal.Lesson
	g.db.First(&lesson, id)
	return lesson, true, nil
}

func (g *GormDb) PutLesson(lesson internal.Lesson) error {
	g.db.Where("lesson_id = ?", lesson.Id).Save(&lesson)
	return nil
}

func (g *GormDb) DeleteLesson(id int) error {
	var lesson internal.Lesson
	g.db.Delete(&lesson, id)
	return nil
}
