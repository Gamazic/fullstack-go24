package main

import (
	"fmt"
	"lessonDb/internal"
	"lessonDb/internal/db/gorm"
)

func main() {
	var db internal.LessonDB
	//db = inmemory.NewInmemoryDb()
	//db = DbStub{}
	db, err := gorm.NewGormDb()
	if err != nil {
		panic(err)
	}

	// Print lessons
	lessons, err := db.GetLessons(100, 0)
	if err != nil {
		panic(err)
	}
	println("All lessons:")
	for _, lesson := range lessons {
		fmt.Printf("%v %v %+v\n", lesson.Id, lesson.Name, lesson.Students)
	}

	// Add two lessons
	println("Add two lessons...")
	lesson1 := internal.Lesson{
		Id:       1,
		Name:     "lesson 1",
		Students: "1,2",
	}
	err = db.PutLesson(lesson1)
	if err != nil {
		panic(err)
	}

	lesson2 := internal.Lesson{
		//Id:       2,
		Name:     "lesson 2",
		Students: "1,2,4,5",
	}
	err = db.PutLesson(lesson2)
	if err != nil {
		panic(err)
	}

	// Print all lessons
	lessons, err = db.GetLessons(100, 0)
	if err != nil {
		panic(err)
	}
	println("All lessons:")
	for _, lesson := range lessons {
		fmt.Printf("%v %v %+v\n", lesson.Id, lesson.Name, lesson.Students)
	}

	// Delete lesson 1
	println("Delete lesson 1...")
	db.DeleteLesson(1)
	// Print all lessons
	lessons, err = db.GetLessons(2, 0)
	if err != nil {
		panic(err)
	}
	println("All lessons:")
	for _, lesson := range lessons {
		fmt.Printf("%v %v %+v\n", lesson.Id, lesson.Name, lesson.Students)
	}
}

//type DbStub struct{}
//
//func (db DbStub) GetLessons() []internal.Lesson {
//	return []internal.Lesson{
//		{1, "lesson 1", "1,2"},
//		{2, "lesson 2", "2,3,4,5"},
//	}
//}
//
//func (db DbStub) GetLesson(id int) internal.Lesson {
//	return internal.Lesson{
//		Id:       1,
//		Name:     "lesson 1",
//		Students: "1,2",
//	}
//}
//
//func (db DbStub) PutLesson(lesson internal.Lesson) {}
//
//func (db DbStub) DeleteLesson(id int) {}
