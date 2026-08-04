package ingest

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// TaskTypeProcess is shared by apps/backend (enqueue) and apps/worker
// (handle) — living here in go-core keeps both sides of the payload
// contract in one place instead of duplicated per app.
const TaskTypeProcess = "ingest:process"

type ProcessPayload struct {
	JobID uuid.UUID `json:"job_id"`
}

func NewProcessTask(jobID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(ProcessPayload{JobID: jobID})
	if err != nil {
		return nil, fmt.Errorf("ingest: marshal task payload: %w", err)
	}
	return asynq.NewTask(TaskTypeProcess, payload), nil
}
