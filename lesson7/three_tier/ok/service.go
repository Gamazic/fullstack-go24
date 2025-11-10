package too_easy

import "fmt"

type NotesService struct {
	notesStorage NotesStorage
}

type NotesStorage interface {
	GetNotes() ([]Note, error)
	SaveNote(note Note) error
	EditNote(id int, note Note) error
	GetNoteById(id int) (Note, error)
	DeleteNote(id int) error
}

func (n *NotesService) EditNote(userId, id int, note Note) error {
	user, err := n.userStorage.Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	allowed := false
	if user.Type == "admin" {
		allowed = true
	}

	note, err = n.notesStorage.GetNoteById(id)
	if err != nil {
		return fmt.Errorf("failed to get note by id: %w", err)
	}
	if user.Id == note.AuthorId {
		allowed = true
	}

	if !allowed {
		return fmt.Errorf("user %s does not have rights to edit note", user)
	}

	return n.notesStorage.SaveNote(id, note)
}
