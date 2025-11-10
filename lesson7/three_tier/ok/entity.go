package too_easy

import "time"

type Note struct {
	Id           int
	Author       string
	Header       string
	Content      string
	CreatedDate  time.Time
	DeadlineDate time.Time
}
