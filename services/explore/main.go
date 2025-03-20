package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"strconv"
	"time"

	pb "github.com/Dev-AustinPeter/explore-service-muzz/genproto/explore"
	_ "github.com/go-sql-driver/mysql" // Register MySQL driver
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	PAGE_LIMIT = 10
)

type server struct {
	pb.UnimplementedExploreServiceServer
	db *sql.DB
}

// ListLikedYou retrieves users who liked the recipient
func (s *server) ListLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	// Default limit per page
	limit := PAGE_LIMIT
	offset := 0
	if req.PaginationToken != nil {
		parsedOffset, err := strconv.Atoi(*req.PaginationToken)
		if err == nil {
			offset = parsedOffset
		}
	}

	query := "SELECT actor_user_id, UNIX_TIMESTAMP(unix_timestamp) FROM decisions WHERE recipient_user_id = ? AND liked = 1 ORDER BY unix_timestamp DESC LIMIT ? OFFSET ?"
	rows, err := s.db.Query(query, req.RecipientUserId, PAGE_LIMIT, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likers []*pb.ListLikedYouResponse_Liker
	for rows.Next() {
		var actorID string
		var timestamp uint64
		if err := rows.Scan(&actorID, &timestamp); err != nil {
			return nil, err
		}
		likers = append(likers, &pb.ListLikedYouResponse_Liker{
			ActorId:       actorID,
			UnixTimestamp: timestamp,
		})
	}

	// Generate the next pagination token
	var nextPaginationToken string
	if len(likers) == limit {
		nextPaginationToken = strconv.Itoa(offset + limit)
	}

	response := &pb.ListLikedYouResponse{
		Likers:              likers,
		NextPaginationToken: &nextPaginationToken,
	}

	return response, nil
}

// ListNewLikedYou retrieves users who liked the recipient but haven’t been liked back
func (s *server) ListNewLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	limit := PAGE_LIMIT
	offset := 0
	if req.PaginationToken != nil {
		parsedOffset, err := strconv.Atoi(*req.PaginationToken)
		if err == nil {
			offset = parsedOffset
		}
	}

	query := `SELECT d1.actor_user_id, UNIX_TIMESTAMP(d1.unix_timestamp) FROM decisions d1 
		WHERE d1.recipient_user_id = ? AND d1.liked = 1 
		AND NOT EXISTS (
			SELECT 1 FROM decisions d2 
			WHERE d2.actor_user_id = d1.recipient_user_id 
			AND d2.recipient_user_id = d1.actor_user_id 
			AND d2.liked = 1
		) 
		ORDER BY d1.unix_timestamp DESC LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, req.RecipientUserId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likers []*pb.ListLikedYouResponse_Liker
	for rows.Next() {
		var actorID string
		var timestamp uint64
		if err := rows.Scan(&actorID, &timestamp); err != nil {
			return nil, err
		}
		likers = append(likers, &pb.ListLikedYouResponse_Liker{
			ActorId:       actorID,
			UnixTimestamp: timestamp,
		})
	}

	var nextPaginationToken string
	if len(likers) == limit {
		nextPaginationToken = strconv.Itoa(offset + limit)
	}

	response := &pb.ListLikedYouResponse{
		Likers:              likers,
		NextPaginationToken: &nextPaginationToken,
	}

	return response, nil
}

// CountLikedYou retrieves the total count of likes received
func (s *server) CountLikedYou(ctx context.Context, req *pb.CountLikedYouRequest) (*pb.CountLikedYouResponse, error) {

	var count uint64
	err := s.db.QueryRow("SELECT COUNT(*) FROM decisions WHERE recipient_user_id = ? AND liked = 1", req.RecipientUserId).Scan(&count)
	if err != nil {
		return nil, err
	}

	return &pb.CountLikedYouResponse{Count: count}, nil
}

// PutDecision with cache invalidation
func (s *server) PutDecision(ctx context.Context, req *pb.PutDecisionRequest) (*pb.PutDecisionResponse, error) {
	// Convert boolean to integer for MySQL
	likedValue := 0
	if req.LikedRecipient {
		likedValue = 1
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO decisions (actor_user_id, recipient_user_id, liked) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE liked = ?", req.ActorUserId, req.RecipientUserId, likedValue, likedValue)
	if err != nil {
		return nil, err
	}

	var mutual bool
	err = tx.QueryRow("SELECT EXISTS (SELECT 1 FROM decisions WHERE actor_user_id = ? AND recipient_user_id = ? AND liked = 1) AND EXISTS (SELECT 1 FROM decisions WHERE actor_user_id = ? AND recipient_user_id = ? AND liked = 1)", req.RecipientUserId, req.ActorUserId, req.ActorUserId, req.RecipientUserId).Scan(&mutual)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &pb.PutDecisionResponse{MutualLikes: mutual}, nil
}

// Retry connecting to the database
func connectToDB(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ { // Retry 10 times
		db, err = sql.Open("mysql", dsn)
		if err == nil && db.Ping() == nil {
			return db, nil
		}
		log.Println("Waiting for database to be ready...")
		time.Sleep(3 * time.Second) // Wait before retrying
	}
	return nil, err
}

func main() {

	// Database connection
	dsn := "user:password@tcp(db:3306)/muzz"

	db, err := connectToDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterExploreServiceServer(s, &server{db: db})

	// Enable gRPC reflection
	reflection.Register(s)

	log.Println("gRPC server is running on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
