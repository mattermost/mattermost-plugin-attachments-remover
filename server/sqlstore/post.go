package sqlstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
)

const (
	postsTable = "Posts"
)

func (s *SQLStore) AttachFileIDsToPost(postID string, fileIDs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	fileIDsJSON, err := json.Marshal(fileIDs)
	if err != nil {
		return fmt.Errorf("error marshaling fileIds: %w", err)
	}

	query := s.masterBuilder.
		Update(postsTable).
		Set("FileIds", fileIDsJSON).
		Where(sq.Eq{"Id": postID})

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
		s.logger.Error("error attaching file to post", "postId", postID, "err", err)
		return tx.Rollback()
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
