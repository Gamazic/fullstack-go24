package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"lessonDb/internal"
)

type MysqlDb struct {
	db *sql.DB
}

func NewMysqlDb() (*MysqlDb, error) {
	db, err := sql.Open("mysql", "root:pass@tcp(localhost:3306)/mydb")
	if err != nil {
		return nil, err
	}
	return &MysqlDb{db: db}, nil
}

func (m *MysqlDb) GetLessons(name string, limit, offset int) ([]internal.Lesson, error) {
	lessons := make([]internal.Lesson, 0)
	rows, err := m.db.Query(
		"SELECT id, name, students FROM lessons LIMIT ? OFFSET ? WHERE name=?;",
		limit, offset, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		l := &internal.Lesson{}
		err = rows.Scan(&l.Id, &l.Name, &l.Students)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, *l)
	}
	return lessons, nil
}

func (m *MysqlDb) GetLessonById(id int) (internal.Lesson, bool, error) {
	return internal.Lesson{}, true, nil
}

func (m *MysqlDb) PutLesson(lesson internal.Lesson) error {
	return nil
}

func (m *MysqlDb) DeleteLesson(id int) error {
	return nil
}
