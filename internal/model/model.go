package model

import "time"

const (
	StatusPending  = "Pending"
	StatusRunning  = "Running"
	StatusComplete = "Complete"
)

type ScheduleRequest struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type Job struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Image  string    `json:"image"`
	Status string    `json:"status"`
	Date   time.Time `json:"date"`
}
