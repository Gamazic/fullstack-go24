package too_easy

import "database/sql"

type StoragePostgres struct {
	db *sql.DB
}

func NewStoragePostgres(db *sql.DB) *StoragePostgres {
	return &StoragePostgres{db: db}
}

func (s *StoragePostgres) GetNotes() ([]NoteModel, error) {
	rows, err := s.db.Query("SELECT id, author, header, content FROM notes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note NoteModel
		err := rows.Scan(&note.Id, &note.Author, &note.Header, &note.Content)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (s *StoragePostgres) SaveNote(note Note) error {
	_, err := s.db.Exec("INSERT INTO notes (author, header, content) VALUES ($1, $2, $3)", note.Author, note.Header, note.Content)
	return err
}
