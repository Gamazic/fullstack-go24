package too_easy

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type NoteModel struct {
	Id             *int    `json:"id"`
	Author         *string `json:"author"`
	Header         *string `json:"header"`
	Content        *string `json:"content"`
	DeadlineInWeek *int    `json:"deadline_in_week"`
}

type NotesApi struct {
	server  http.Server
	service NotesService
}

func (a *NotesApi) GetNoteById(w http.ResponseWriter, r *http.Request) {
	idRaw := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idRaw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	note, err := a.service.GetNoteById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	noteResponse := NoteModel{
		Id:      &note.Id,
		Author:  &note.Author,
		Header:  &note.Header,
		Content: &note.Content,
		DeadlineInWeek: (note.Deadline - note.Created).week
	}

	jsonNote, err := json.Marshal(noteResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonNote)
}
