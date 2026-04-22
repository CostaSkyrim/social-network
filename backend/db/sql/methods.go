package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

//====================================
// MODELS
//====================================

type User struct {
	ID           int64
	UUID         string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Nickname     string
	DateOfBirth  string
	AboutMe      string
	AvatarPath   string
	IsPublic     bool
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Session struct {
	ID        int64
	SessionID string
	UserID    int64
	IPAddress string
	UserAgent string
	ExpiresAt time.Time
	IsActive  bool
	CreatedAt time.Time
}

type Post struct {
	ID           int64
	UUID         string
	AuthorID     int64
	GroupID      int64
	Content      string
	ImagePath    string
	PrivacyLevel string
	IsDeleted    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Group struct {
	ID            int64
	UUID          string
	CreatorID     int64
	Title         string
	Description   string
	AvatarPath    string
	LastMessageAt time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type GroupMember struct {
	ID        int64
	GroupID   int64
	UserID    int64
	Status    string
	InvitedBy int64
	JoinedAt  time.Time
	UpdatedAt time.Time
}

type Message struct {
	ID              int64
	UUID            string
	SenderID        int64
	DirectMessageID int64
	GroupID         int64
	Content         string
	IsRead          bool
	CreatedAt       time.Time
}

type DirectMessage struct {
	ID            int64
	User1ID       int64
	User2ID       int64
	LastMessageAt time.Time
	CreatedAt     time.Time
}

type Notification struct {
	ID         int64
	UserID     int64
	FromUserID int64
	Type       string
	Content    string
	RelatedID  int64
	IsRead     bool
	ReadAt     time.Time
	CreatedAt  time.Time
}

//====================================
// USER METHODS
//====================================

// AddUser adds a new user to the database
func (db *DataBase) AddUser(ctx context.Context, user *User) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := db.conn.BeginTx(dbCtx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if user.PasswordHash == "" {
		return 0, fmt.Errorf("password hash cannot be empty")
	}

	result, err := tx.ExecContext(dbCtx,
		`INSERT INTO users (uuid, email, password_hash, first_name, last_name, 
		 nickname, date_of_birth, about_me, avatar_path, is_public)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.UUID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Nickname,
		user.DateOfBirth,
		user.AboutMe,
		user.AvatarPath,
		user.IsPublic,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// GetUserByEmail retrieves a user by email
func (db *DataBase) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	user := &User{}
	err := db.conn.QueryRowContext(ctx,
		`SELECT id, uuid, email, password_hash, first_name, last_name, nickname,
		 date_of_birth, about_me, avatar_path, is_public, is_active, created_at, updated_at
		 FROM users WHERE email = ? AND is_active = 1`,
		email,
	).Scan(
		&user.ID,
		&user.UUID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Nickname,
		&user.DateOfBirth,
		&user.AboutMe,
		&user.AvatarPath,
		&user.IsPublic,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (db *DataBase) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	user := &User{}
	err := db.conn.QueryRowContext(ctx,
		`SELECT id, uuid, email, first_name, last_name, nickname,
		 date_of_birth, about_me, avatar_path, is_public, created_at
		 FROM users WHERE id = ? AND is_active = 1`,
		userID,
	).Scan(
		&user.ID,
		&user.UUID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Nickname,
		&user.DateOfBirth,
		&user.AboutMe,
		&user.AvatarPath,
		&user.IsPublic,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

// GetUserByUUID retrieves a user by UUID
func (db *DataBase) GetUserByUUID(ctx context.Context, uuid string) (*User, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	user := &User{}
	err := db.conn.QueryRowContext(ctx,
		`SELECT id, uuid, email, first_name, last_name, nickname,
		 date_of_birth, about_me, avatar_path, is_public, created_at
		 FROM users WHERE uuid = ? AND is_active = 1`,
		uuid,
	).Scan(
		&user.ID,
		&user.UUID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Nickname,
		&user.DateOfBirth,
		&user.AboutMe,
		&user.AvatarPath,
		&user.IsPublic,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

// UpdateUserProfile updates user profile information
func (db *DataBase) UpdateUserProfile(ctx context.Context, userID int64, user *User) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		`UPDATE users SET nickname = COALESCE(?, nickname),
		 about_me = COALESCE(?, about_me),
		 avatar_path = COALESCE(?, avatar_path),
		 updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		user.Nickname,
		user.AboutMe,
		user.AvatarPath,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

// UpdateUserPrivacy updates user privacy settings
func (db *DataBase) UpdateUserPrivacy(ctx context.Context, userID int64, isPublic bool) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		`UPDATE users SET is_public = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		isPublic,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user privacy: %w", err)
	}

	return nil
}

// DeleteUser soft deletes a user
func (db *DataBase) DeleteUser(ctx context.Context, userID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		`UPDATE users SET is_active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

//====================================
// SESSION METHODS
//====================================

// CreateSession creates a new session
func (db *DataBase) CreateSession(ctx context.Context, session *Session) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		`INSERT INTO sessions (session_id, user_id, ip_address, user_agent, expires_at)
		 VALUES (?, ?, ?, ?, ?)`,
		session.SessionID,
		session.UserID,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get session id: %w", err)
	}

	return id, nil
}

// GetSession retrieves a session by session ID
func (db *DataBase) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	session := &Session{}
	err := db.conn.QueryRowContext(ctx,
		`SELECT id, session_id, user_id, expires_at, is_active
		 FROM sessions WHERE session_id = ? AND expires_at > CURRENT_TIMESTAMP AND is_active = 1`,
		sessionID,
	).Scan(
		&session.ID,
		&session.SessionID,
		&session.UserID,
		&session.ExpiresAt,
		&session.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	return session, nil
}

// DeleteSession soft deletes a session
func (db *DataBase) DeleteSession(ctx context.Context, sessionID string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		`UPDATE sessions SET is_active = 0 WHERE session_id = ?`,
		sessionID,
	)

	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

//====================================
// POST METHODS
//====================================

// CreatePost creates a new post
func (db *DataBase) CreatePost(ctx context.Context, post *Post) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		`INSERT INTO posts (uuid, author_id, group_id, content, image_path, privacy_level)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		post.UUID,
		post.AuthorID,
		post.GroupID,
		post.Content,
		post.ImagePath,
		post.PrivacyLevel,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create post: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get post id: %w", err)
	}

	return id, nil
}

// GetPost retrieves a post by ID
func (db *DataBase) GetPost(ctx context.Context, postID, userID int64) (*Post, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	post := &Post{}
	err := db.conn.QueryRowContext(ctx,
		`SELECT id, uuid, author_id, group_id, content, image_path,
		 privacy_level, created_at, updated_at
		 FROM posts WHERE id = ? AND (group_id IS NULL OR group_id IN (SELECT group_id 
			FROM group_members WHERE user_id = ? AND status = 'accepted'))`,
		postID, userID,
	).Scan(
		&post.ID,
		&post.UUID,
		&post.AuthorID,
		&post.GroupID,
		&post.Content,
		&post.ImagePath,
		&post.PrivacyLevel,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to query post: %w", err)
	}

	return post, nil
}

// GetUserPosts retrieves user's posts with pagination
func (db *DataBase) GetUserPosts(ctx context.Context, userID int64, limit, offset int) ([]*Post, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx,
		`SELECT id, uuid, content, image_path, privacy_level, created_at, updated_at
		 FROM posts WHERE author_id = ? AND group_id IS NULL
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query user posts: %w", err)
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(
			&post.ID,
			&post.UUID,
			&post.Content,
			&post.ImagePath,
			&post.PrivacyLevel,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return posts, nil
}

// DeletePost soft deletes a post
func (db *DataBase) DeletePost(ctx context.Context, postID, userID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		`UPDATE posts SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND author_id = ?`,
		postID, userID,
	)

	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}

//====================================
// GROUP METHODS
//====================================

// CreateGroup creates a new group
func (db *DataBase) CreateGroup(ctx context.Context, group *Group) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		`INSERT INTO groups (uuid, creator_id, title, description, avatar_path)
		 VALUES (?, ?, ?, ?, ?)`,
		group.UUID,
		group.CreatorID,
		group.Title,
		group.Description,
		group.AvatarPath,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create group: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get group id: %w", err)
	}

	return id, nil
}

// GetGroup retrieves a group by ID
func (db *DataBase) GetGroup(ctx context.Context, groupID int64) (*Group, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	group := &Group{}
	err := db.conn.QueryRowContext(ctx,
		`SELECT id, uuid, creator_id, title, description, avatar_path,
		 created_at, updated_at, last_message_at
		 FROM groups WHERE id = ?`,
		groupID,
	).Scan(
		&group.ID,
		&group.UUID,
		&group.CreatorID,
		&group.Title,
		&group.Description,
		&group.AvatarPath,
		&group.CreatedAt,
		&group.UpdatedAt,
		&group.LastMessageAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("group not found")
		}
		return nil, fmt.Errorf("failed to query group: %w", err)
	}

	return group, nil
}

// GetUserGroups retrieves all groups a user is member of
func (db *DataBase) GetUserGroups(ctx context.Context, userID int64) ([]*Group, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx,
		`SELECT g.id, g.uuid, g.title, g.description, g.avatar_path, g.last_message_at
		 FROM groups g
		 JOIN group_members gm ON gm.group_id = g.id
		 WHERE gm.user_id = ? AND gm.status = 'accepted'
		 ORDER BY g.last_message_at DESC`,
		userID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query user groups: %w", err)
	}
	defer rows.Close()

	var groups []*Group
	for rows.Next() {
		group := &Group{}
		err := rows.Scan(
			&group.ID,
			&group.UUID,
			&group.Title,
			&group.Description,
			&group.AvatarPath,
			&group.LastMessageAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return groups, nil
}

//====================================
// MESSAGE METHODS
//====================================

// CreateMessage creates a new message
func (db *DataBase) CreateMessage(ctx context.Context, message *Message) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		`INSERT INTO messages (uuid, sender_id, direct_message_id, group_id, content)
		 VALUES (?, ?, ?, ?, ?)`,
		message.UUID,
		message.SenderID,
		message.DirectMessageID,
		message.GroupID,
		message.Content,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get message id: %w", err)
	}

	return id, nil
}

// CreateOrGetDirectMessage creates or gets an existing direct message conversation
func (db *DataBase) CreateOrGetDirectMessage(ctx context.Context, dm *DirectMessage) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// First try to get existing DM
	var dmID int64
	err := db.conn.QueryRowContext(dbCtx,
		`SELECT id FROM direct_messages
		 WHERE (user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)`,
		dm.User1ID, dm.User2ID, dm.User2ID, dm.User1ID,
	).Scan(&dmID)

	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query direct message: %w", err)
	}

	if err == sql.ErrNoRows {
		// Create new DM
		result, err := db.conn.ExecContext(dbCtx,
			`INSERT INTO direct_messages (user1_id, user2_id)
			 VALUES (?, ?)`,
			dm.User1ID, dm.User2ID,
		)

		if err != nil {
			return 0, fmt.Errorf("failed to create direct message: %w", err)
		}

		dmID, err = result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get dm id: %w", err)
		}
	}

	return dmID, nil
}

//====================================
// NOTIFICATION METHODS
//====================================

// CreateNotification creates a new notification
func (db *DataBase) CreateNotification(ctx context.Context, notification *Notification) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		`INSERT INTO notifications (user_id, from_user_id, type, content, related_id)
		 VALUES (?, ?, ?, ?, ?)`,
		notification.UserID,
		notification.FromUserID,
		notification.Type,
		notification.Content,
		notification.RelatedID,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create notification: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get notification id: %w", err)
	}

	return id, nil
}

// GetUserNotifications retrieves user's notifications with pagination
func (db *DataBase) GetUserNotifications(ctx context.Context, userID int64, limit, offset int) ([]*Notification, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx,
		`SELECT id, from_user_id, type, content, is_read, related_id, created_at
		 FROM notifications WHERE user_id = ?
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		notif := &Notification{}
		err := rows.Scan(
			&notif.ID,
			&notif.FromUserID,
			&notif.Type,
			&notif.Content,
			&notif.IsRead,
			&notif.RelatedID,
			&notif.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, notif)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return notifications, nil
}

// MarkNotificationAsRead marks a notification as read
func (db *DataBase) MarkNotificationAsRead(ctx context.Context, notificationID, userID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		`UPDATE notifications SET is_read = 1, read_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND user_id = ?`,
		notificationID, userID,
	)

	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}
