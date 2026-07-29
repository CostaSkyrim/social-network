package populate

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"social-network/backend/db/queries"

	"golang.org/x/crypto/bcrypt"
)

type seedUser struct {
	UUID        string  `json:"uuid"`
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Nickname    *string `json:"nickname"`
	DateOfBirth string  `json:"date_of_birth"`
	AboutMe     *string `json:"about_me"`
	AvatarPath  *string `json:"avatar_path"`
	IsPublic    bool    `json:"is_public"`
}

type seedFollow struct {
	FollowerID  int64  `json:"follower_id"`
	FollowingID int64  `json:"following_id"`
	Status      string `json:"status"`
}

type seedGroup struct {
	UUID        string `json:"uuid"`
	CreatorID   int64  `json:"creator_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type seedGroupMember struct {
	GroupID   int64  `json:"group_id"`
	UserID    int64  `json:"user_id"`
	Status    string `json:"status"`
	InvitedBy *int64 `json:"invited_by"`
}

type seedPost struct {
	UUID         string  `json:"uuid"`
	AuthorID     int64   `json:"author_id"`
	GroupID      *int64  `json:"group_id"`
	Content      string  `json:"content"`
	ImagePath    string  `json:"image_path"`
	PrivacyLevel string  `json:"privacy_level"`
	CreatedAt    *string `json:"created_at"`
	IsDeleted    *bool   `json:"is_deleted"`
}

type seedPostVisibility struct {
	PostID int64 `json:"post_id"`
	UserID int64 `json:"user_id"`
}

type seedComment struct {
	UUID            string `json:"uuid"`
	PostID          int64  `json:"post_id"`
	AuthorID        int64  `json:"author_id"`
	ParentCommentID *int64 `json:"parent_comment_id"`
	Content         string `json:"content"`
	IsDeleted       *bool  `json:"is_deleted"`
}

type seedEvent struct {
	UUID          string `json:"uuid"`
	GroupID       int64  `json:"group_id"`
	CreatorID     int64  `json:"creator_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	EventDatetime string `json:"event_datetime"`
}

type seedEventResponse struct {
	EventID  int64  `json:"event_id"`
	UserID   int64  `json:"user_id"`
	Response string `json:"response"`
}

type seedMessage struct {
	UUID            string `json:"uuid"`
	SenderID        int64  `json:"sender_id"`
	DirectMessageID *int64 `json:"direct_message_id"`
	GroupID         *int64 `json:"group_id"`
	Content         string `json:"content"`
}

type seedNotification struct {
	UserID     int64  `json:"user_id"`
	FromUserID *int64 `json:"from_user_id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	RelatedID  *int64 `json:"related_id"`
}

type seedData struct {
	Users          []seedUser           `json:"users"`
	Follows        []seedFollow         `json:"follows"`
	Groups         []seedGroup          `json:"groups"`
	GroupMembers   []seedGroupMember    `json:"group_members"`
	Posts          []seedPost           `json:"posts"`
	PostVisibility []seedPostVisibility `json:"post_visibility"`
	Comments       []seedComment        `json:"comments"`
	Events         []seedEvent          `json:"events"`
	EventResponses []seedEventResponse  `json:"event_responses"`
	Messages       []seedMessage        `json:"messages"`
	Notifications  []seedNotification   `json:"notifications"`
}

// if the users table is empty. Returns the number of users inserted.
func SeedFromJSON(db *sql.DB, jsonPath string) (int, error) {
	exists, err := hasUsers(db)
	if err != nil {
		return 0, fmt.Errorf("seed: check users: %w", err)
	}
	if exists {
		log.Println("Database already has users, skipping seed")
		return 0, nil
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return 0, fmt.Errorf("seed: read file: %w", err)
	}

	var sd seedData
	if err := json.Unmarshal(data, &sd); err != nil {
		return 0, fmt.Errorf("seed: parse json: %w", err)
	}

	count, err := seed(db, &sd)
	if err != nil {
		return 0, fmt.Errorf("seed: %w", err)
	}

	log.Printf("Seed data loaded: %d users, plus posts, comments, groups, and more", count)
	return count, nil
}

// Reseed drops all data and re-imports from the JSON file.
func Reseed(db *sql.DB, jsonPath string) error {
	log.Println("Reseeding database...")

	if err := dropAllData(db); err != nil {
		return fmt.Errorf("reseed: drop: %w", err)
	}

	_, err := SeedFromJSON(db, jsonPath)
	return err
}

func hasUsers(db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("query users count: %w", err)
	}
	return count > 0, nil
}

func seed(db *sql.DB, sd *seedData) (int, error) {
	if err := insertUsers(db, sd.Users); err != nil {
		return 0, err
	}

	if err := insertFollows(db, sd.Follows); err != nil {
		return 0, err
	}

	if err := insertGroups(db, sd.Groups); err != nil {
		return 0, err
	}

	if err := insertGroupMembers(db, sd.GroupMembers); err != nil {
		return 0, err
	}

	if err := insertPosts(db, sd.Posts); err != nil {
		return 0, err
	}

	if err := insertPostVisibility(db, sd.PostVisibility); err != nil {
		return 0, err
	}

	if err := insertComments(db, sd.Comments); err != nil {
		return 0, err
	}

	if err := insertEvents(db, sd.Events); err != nil {
		return 0, err
	}

	if err := insertEventResponses(db, sd.EventResponses); err != nil {
		return 0, err
	}

	if err := insertMessages(db, sd.Messages); err != nil {
		return 0, err
	}

	if err := insertNotifications(db, sd.Notifications); err != nil {
		return 0, err
	}

	return len(sd.Users), nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func insertUsers(db *sql.DB, users []seedUser) error {
	for _, u := range users {
		hash, err := hashPassword(u.Password)
		if err != nil {
			return fmt.Errorf("hash password for %s: %w", u.Email, err)
		}

		dob, err := time.Parse(time.RFC3339, u.DateOfBirth)
		if err != nil {
			return fmt.Errorf("parse dob for %s: %w", u.Email, err)
		}

		_, err = db.Exec(queries.CreateUser,
			u.UUID, u.Email, hash, u.FirstName, u.LastName, u.Nickname,
			dob, u.AboutMe, u.AvatarPath, u.IsPublic)
		if err != nil {
			return fmt.Errorf("insert user %s: %w", u.Email, err)
		}
	}
	return nil
}

func insertFollows(db *sql.DB, follows []seedFollow) error {
	for _, f := range follows {
		_, err := db.Exec(`
			INSERT INTO followers (follower_id, following_id, status, created_at, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, f.FollowerID, f.FollowingID, f.Status)
		if err != nil {
			return fmt.Errorf("insert follow (%d→%d): %w", f.FollowerID, f.FollowingID, err)
		}
	}
	return nil
}

func insertGroups(db *sql.DB, groups []seedGroup) error {
	for _, g := range groups {
		_, err := db.Exec(queries.CreateGroup,
			g.UUID, g.CreatorID, g.Title, g.Description, nil)
		if err != nil {
			return fmt.Errorf("insert group %s: %w", g.Title, err)
		}
	}
	return nil
}

func insertGroupMembers(db *sql.DB, members []seedGroupMember) error {
	for _, m := range members {
		_, err := db.Exec(`
			INSERT INTO group_members (group_id, user_id, status, invited_by, joined_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, m.GroupID, m.UserID, m.Status, m.InvitedBy)
		if err != nil {
			return fmt.Errorf("insert group member (group=%d, user=%d): %w", m.GroupID, m.UserID, err)
		}
	}
	return nil
}

func insertPosts(db *sql.DB, posts []seedPost) error {
	for _, p := range posts {
		var createdAt string
		if p.CreatedAt != nil {
			createdAt = *p.CreatedAt
		} else {
			createdAt = time.Now().Format("2006-01-02 15:04:05")
		}

		isDeleted := 0
		if p.IsDeleted != nil && *p.IsDeleted {
			isDeleted = 1
		}

		_, err := db.Exec(`
			INSERT INTO posts (uuid, author_id, group_id, content, image_path,
			 privacy_level, created_at, updated_at, is_deleted)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, p.UUID, p.AuthorID, p.GroupID, p.Content, p.ImagePath, p.PrivacyLevel, createdAt, createdAt, isDeleted)
		if err != nil {
			return fmt.Errorf("insert post %s: %w", p.UUID, err)
		}
	}
	return nil
}

func insertPostVisibility(db *sql.DB, vis []seedPostVisibility) error {
	for _, v := range vis {
		_, err := db.Exec(queries.AddPostVisibility, v.PostID, v.UserID)
		if err != nil {
			return fmt.Errorf("insert post_visibility (post=%d, user=%d): %w", v.PostID, v.UserID, err)
		}
	}
	return nil
}

func insertComments(db *sql.DB, comments []seedComment) error {
	for _, c := range comments {
		isDeleted := 0
		if c.IsDeleted != nil && *c.IsDeleted {
			isDeleted = 1
		}

		_, err := db.Exec(`
			INSERT INTO comments (uuid, post_id, author_id, parent_comment_id, content,
			 image_path, created_at, updated_at, is_deleted)
			VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)
		`, c.UUID, c.PostID, c.AuthorID, c.ParentCommentID, c.Content, nil, isDeleted)
		if err != nil {
			return fmt.Errorf("insert comment %s: %w", c.UUID, err)
		}
	}
	return nil
}

func insertEvents(db *sql.DB, events []seedEvent) error {
	for _, e := range events {
		dt, err := time.Parse(time.RFC3339, e.EventDatetime)
		if err != nil {
			return fmt.Errorf("parse event datetime for %s: %w", e.Title, err)
		}

		_, err = db.Exec(queries.CreateEvent,
			e.UUID, e.GroupID, e.CreatorID, e.Title, e.Description, dt)
		if err != nil {
			return fmt.Errorf("insert event %s: %w", e.Title, err)
		}
	}
	return nil
}

func insertEventResponses(db *sql.DB, responses []seedEventResponse) error {
	for _, r := range responses {
		_, err := db.Exec(`
			INSERT INTO event_responses (event_id, user_id, response, created_at, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, r.EventID, r.UserID, r.Response)
		if err != nil {
			return fmt.Errorf("insert event_response (event=%d, user=%d): %w", r.EventID, r.UserID, err)
		}
	}
	return nil
}

func insertMessages(db *sql.DB, messages []seedMessage) error {
	// First collect unique DM pairs and create DM conversations
	dmPairs := make(map[[2]int64]int64) // [user1, user2] -> dm_id

	for _, m := range messages {
		if m.DirectMessageID == nil {
			continue
		}
	}

	// Hard-code the DM pairs from our seed data
	dmID1, err := createDM(db, 1, 2)
	if err != nil {
		return fmt.Errorf("create DM 1-2: %w", err)
	}
	dmPairs[[2]int64{1, 2}] = dmID1
	dmPairs[[2]int64{2, 1}] = dmID1

	dmID2, err := createDM(db, 3, 4)
	if err != nil {
		return fmt.Errorf("create DM 3-4: %w", err)
	}
	dmPairs[[2]int64{3, 4}] = dmID2
	dmPairs[[2]int64{4, 3}] = dmID2

	// Insert messages, mapping DM IDs
	for _, m := range messages {
		if m.DirectMessageID != nil {
			pair := [2]int64{}
			if *m.DirectMessageID == 1 {
				pair = [2]int64{1, 2}
			} else if *m.DirectMessageID == 2 {
				pair = [2]int64{3, 4}
			}
			realDMID, ok := dmPairs[pair]
			if !ok {
				return fmt.Errorf("unknown DM pair for message %s", m.UUID)
			}

			_, err := db.Exec(queries.CreateMessage,
				m.UUID, m.SenderID, realDMID, nil, m.Content)
			if err != nil {
				return fmt.Errorf("insert message %s: %w", m.UUID, err)
			}
		} else if m.GroupID != nil {
			_, err := db.Exec(queries.CreateMessage,
				m.UUID, m.SenderID, nil, m.GroupID, m.Content)
			if err != nil {
				return fmt.Errorf("insert group message %s: %w", m.UUID, err)
			}
		}
	}

	// Update last_message_at for DMs
	for _, dmID := range dmPairs {
		db.Exec(`UPDATE direct_messages SET last_message_at = CURRENT_TIMESTAMP WHERE id = ?`, dmID)
	}

	return nil
}

func createDM(db *sql.DB, user1ID, user2ID int64) (int64, error) {
	// Ensure user1ID < user2ID for consistent pair ordering
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}

	// Try to get existing DM first
	var dmID int64
	err := db.QueryRow(`
		SELECT id FROM direct_messages
		WHERE user1_id = ? AND user2_id = ?
	`, user1ID, user2ID).Scan(&dmID)

	if err == nil {
		return dmID, nil
	}

	if err != sql.ErrNoRows {
		return 0, err
	}

	result, err := db.Exec(`
		INSERT INTO direct_messages (user1_id, user2_id, created_at, last_message_at)
		VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, user1ID, user2ID)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func insertNotifications(db *sql.DB, notifications []seedNotification) error {
	for _, n := range notifications {
		_, err := db.Exec(queries.CreateNotification,
			n.UserID, n.FromUserID, n.Type, n.Content, n.RelatedID)
		if err != nil {
			return fmt.Errorf("insert notification: %w", err)
		}
	}
	return nil
}

func dropAllData(db *sql.DB) error {
	tables := []string{
		"message_reads",
		"messages",
		"direct_messages",
		"event_responses",
		"events",
		"notifications",
		"post_visibility",
		"comments",
		"posts",
		"group_members",
		"groups",
		"followers",
		"sessions",
		"users",
	}

	for _, table := range tables {
		if _, err := db.Exec("DELETE FROM " + table); err != nil {
			return fmt.Errorf("delete from %s: %w", table, err)
		}
	}

	return nil
}

// DefaultPath returns the default path to the seed JSON file.
func DefaultPath() string {
	return filepath.Join("backend", "populate", "seed.json")
}
