package too_easy

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type NoteModel struct {
	Id      *int    `json:"id"`
	Author  *string `json:"author"`
	Header  *string `json:"header"`
	Content *string `json:"content"`
}

type NotesApi struct {
	server  http.Server
	storage NotesStorage
}

type NotesStorage interface {
	GetNotes() ([]Note, error)
	SaveNote(note Note) error
	EditNote(id int, note Note) error
	GetNoteById(id int) (Note, error)
	DeleteNote(id int) error
}

func (a *NotesApi) GetNoteById(w http.ResponseWriter, r *http.Request) {
	idRaw := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idRaw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	noteResponse, err := a.storage.GetNoteById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	jsonNote, err := json.Marshal(noteResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonNote)
}
