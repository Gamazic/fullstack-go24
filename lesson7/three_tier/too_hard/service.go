package too_easy

import (
	"encoding/json"
	"fmt"
)

type NotesService struct {
	notesStorage NotesStorage
}

func (n *NotesService) EditNote(userId, id int, note Note, otherData OtherData) (SpecialStatuses, error) {
	user, err := n.userStorage.Get(userId)
	if err != nil {
		status := SpecialStatusses{
			Status:    StatusFailed,
			Reason:    fmt.Sprintf("failed to get user: %v", err),
			User:      user,
			Note:      note,
			OtherData: otherData,
		}
		return status, fmt.Errorf("failed to get user: %w", err)
	}
	note, err := n.notesStorage.GetNoteById(id)
	if err != nil {
		status := SpecialStatusses{
			Status:    StatusFailedOnNote,
			Reason:    fmt.Sprintf("failed to get note by id: %v", err),
			User:      user,
			Note:      note,
			OtherData: otherData,
		}
		return someStatus, fmt.Errorf("failed to get note by id: %w", err)
	}
	var allowed bool
	if user.Type == "admin" {
		allowed = true
	}
	if user.Type == "user" {
		if note.Author.Id == user.Id {
			allowed = true
		}
	}
	if otherData.SuperPriviledgeV2 {
		superPriveledgeId := n.priveledgeStorage.GetSuperPriveledgeId(note.id)
		if user.HasRightToBeSuperPriviledged.Id == superPriveledgeId {
			allowed = true
		}
	}
	if !allowed {
		data, err := json.Marshal(map[string]any{"note": note, "user": user})
		if err != nil {
			// form super status
			return someStatus, fmt.Errorf("failed to marshal data: %w", err)
		}
		n.kafkaClient.SendEvent(data, eventEditNoteNotAllowed)
	}
	if allowed {
		data, err := json.Marshal(map[string]any{"note": note, "user": user})
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		n.kafkaClient.SendEven(note, user, evenEditNoteAllowed)
	}

	if otherData.Version == versionV3Fix {
		note.Content = note.Content + "really important change here for specific version"
	}

	// ok we checked rights, now let's change note

	// if note modified 3 years ago, then send notification
	// if field content has word "peace" in it, then send notification, delete note delete user
	// ................
}
