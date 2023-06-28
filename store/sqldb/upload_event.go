package sqldb

import (
	"fmt"
	"time"
)

func (s *SpDBImpl) InsertUploadEvent(objectID uint64, state string, description string) error {
	updateTime := time.Now().String()
	if result := s.db.Create(&UploadEventTable{
		ObjectID:    objectID,
		UploadState: state,
		Description: description,
		UpdateTime:  updateTime,
	}); result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to insert upload event record: %s", result.Error)
	}
	return nil
}
