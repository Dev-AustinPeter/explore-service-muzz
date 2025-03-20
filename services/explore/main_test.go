//go:build unit

package main

/*
 * @author Austin Rodrigues
 * @FilePath: /explore-service-muzz/services/explore/main_test.go
 * @Description: gRPC server for Muzz's Explore Service
 */
import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	pb "github.com/Dev-AustinPeter/explore-service-muzz/genproto/explore"
)

func TestListLikedYou(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing mock database: %v", err)
	}
	defer db.Close()

	s := &server{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT actor_user_id, UNIX_TIMESTAMP(unix_timestamp) FROM decisions WHERE recipient_user_id = ? AND liked = 1 ORDER BY unix_timestamp DESC LIMIT ? OFFSET ?`)).
		WithArgs("user2", 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"actor_user_id", "unix_timestamp"}).AddRow("user1", 1710456789))

	req := &pb.ListLikedYouRequest{RecipientUserId: "user2"}
	resp, err := s.ListLikedYou(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(resp.Likers) != 1 || resp.Likers[0].ActorId != "user1" {
		t.Fatalf("Unexpected response: %+v", resp)
	}
}

func TestCountLikedYou(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing mock database: %v", err)
	}
	defer db.Close()

	s := &server{db: db}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM decisions WHERE recipient_user_id = ? AND liked = 1`)).
		WithArgs("user2").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	req := &pb.CountLikedYouRequest{RecipientUserId: "user2"}
	resp, err := s.CountLikedYou(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.Count != 10 {
		t.Fatalf("Expected count 10, got %d", resp.Count)
	}
}

func TestListNewLikedYou(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing mock database: %v", err)
	}
	defer db.Close()

	s := &server{db: db}
	expectedQuery := `SELECT d1.actor_user_id, UNIX_TIMESTAMP(d1.unix_timestamp) FROM decisions d1 
		WHERE d1.recipient_user_id = ? AND d1.liked = 1 
		AND NOT EXISTS (
			SELECT 1 FROM decisions d2 
			WHERE d2.actor_user_id = d1.recipient_user_id 
			AND d2.recipient_user_id = d1.actor_user_id 
			AND d2.liked = 1
		) 
		ORDER BY d1.unix_timestamp DESC LIMIT ? OFFSET ?`

	mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
		WithArgs("user2", 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"actor_user_id", "unix_timestamp"}).AddRow("user1", 1710456789))

	req := &pb.ListLikedYouRequest{RecipientUserId: "user2"}
	resp, err := s.ListNewLikedYou(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(resp.Likers) != 1 || resp.Likers[0].ActorId != "user1" {
		t.Fatalf("Unexpected response: %+v", resp)
	}
}

func TestPutDecision(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing mock database: %v", err)
	}
	defer db.Close()

	svc := &server{db: db}

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO decisions (actor_user_id, recipient_user_id, liked) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE liked = ?`)).
		WithArgs("user1", "user2", 1, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS (SELECT 1 FROM decisions WHERE actor_user_id = ? AND recipient_user_id = ? AND liked = 1) AND EXISTS (SELECT 1 FROM decisions WHERE actor_user_id = ? AND recipient_user_id = ? AND liked = 1)`)).
		WithArgs("user2", "user1", "user1", "user2").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(0))

	mock.ExpectCommit()

	req := &pb.PutDecisionRequest{ActorUserId: "user1", RecipientUserId: "user2", LikedRecipient: true}
	resp, err := svc.PutDecision(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.MutualLikes {
		t.Fatalf("Expected mutual likes to be false, got true")
	}
}
