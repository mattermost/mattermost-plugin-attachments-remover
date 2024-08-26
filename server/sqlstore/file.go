package sqlstore

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/mattermost/mattermost/server/public/model"
)

const (
	fileInfoTable = "FileInfo"
)

func (s *SQLStore) DetachAttachmentFromChannel(fileID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	update := map[string]interface{}{
		"ChannelId": "",
		"DeleteAt":  model.GetMillis(),
	}

	query := s.masterBuilder.
		Update(fileInfoTable).
		SetMap(update).
		Where(sq.Eq{"Id": fileID})

	tx, err := s.master.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	q, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error creating query: %w", err)
	}

	_, err = tx.ExecContext(ctx, q, args...)
	if err != nil {
		s.logger.Error("error detaching attachment from channel", "fileId", fileID, "err", err)
		if errRollback := tx.Rollback(); errRollback != nil {
			s.logger.Error("error rolling back transaction", "err", errRollback)
		}
		return fmt.Errorf("error detaching attachment from channel: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
