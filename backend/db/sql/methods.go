package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"social-network/backend/db/queries"
)

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
		queries.CreateUser,
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
		queries.GetUserByEmail,
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

// GetUserByNickname retrieves a user by nickname
func (db *DataBase) GetUserByNickname(ctx context.Context, nickname string) (*User, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	user := &User{}
	err := db.conn.QueryRowContext(ctx,
		queries.GetUserByNickname,
		nickname,
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
		queries.GetUserByID,
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
		queries.GetUserByUUID,
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
		queries.UpdateUserProfile,
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

// UpdateUserAvatar updates a user's avatar path
func (db *DataBase) UpdateUserAvatar(ctx context.Context, userID int64, avatarPath string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		queries.UpdateUserAvatar,
		avatarPath,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user avatar: %w", err)
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
		queries.UpdateUserPrivacy,
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
		queries.DeleteUser,
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
		queries.CreateSession,
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
		queries.GetSession,
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
		queries.DeleteSession,
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (db *DataBase) DeleteAllUserSessions(ctx context.Context, userID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx, queries.DeleteAllUserSessions, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

//====================================
// FOLLOW METHODS
//====================================

// CreateFollowRequest sends a follow request
func (db *DataBase) CreateFollowRequest(ctx context.Context, followerID, followingID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		queries.CreateFollowRequest,
		followerID, followingID,
	)
	if err != nil {
		return fmt.Errorf("failed to create follow request: %w", err)
	}

	return nil
}

// AcceptFollowRequest accepts a follow request
func (db *DataBase) AcceptFollowRequest(ctx context.Context, followerID, followingID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		queries.UpdateFollowStatus,
		"accepted", followerID, followingID,
	)
	if err != nil {
		return fmt.Errorf("failed to accept follow request: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("follow request not found")
	}

	return nil
}

// DeclineFollowRequest declines a follow request
func (db *DataBase) DeclineFollowRequest(ctx context.Context, followerID, followingID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		queries.UpdateFollowStatus,
		"declined", followerID, followingID,
	)
	if err != nil {
		return fmt.Errorf("failed to decline follow request: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("follow request not found")
	}

	return nil
}

// RemoveFollow removes a follow relationship
func (db *DataBase) RemoveFollow(ctx context.Context, followerID, followingID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		`DELETE FROM followers WHERE follower_id = ? AND following_id = ?`,
		followerID, followingID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove follow: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("follow relationship not found")
	}

	return nil
}

// GetFollowers retrieves the followers of a user
func (db *DataBase) GetFollowers(ctx context.Context, userID int64) ([]User, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx, queries.GetFollowers, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query followers: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(
			&u.ID, &u.UUID, &u.Email, &u.FirstName, &u.LastName,
			&u.Nickname, &u.AvatarPath, &u.IsPublic,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan follower: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return users, nil
}

// GetFollowing retrieves the users a user is following
func (db *DataBase) GetFollowing(ctx context.Context, userID int64) ([]User, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx, queries.GetFollowing, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query following: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(
			&u.ID, &u.UUID, &u.Email, &u.FirstName, &u.LastName,
			&u.Nickname, &u.AvatarPath, &u.IsPublic,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan following user: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return users, nil
}

// GetFollowingWithDM retrieves the users a user is following with DM unread counts
func (db *DataBase) GetFollowingWithDM(ctx context.Context, userID int64) ([]FollowerWithDM, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx, queries.GetFollowingWithDM, userID, userID, userID)
	if err != nil {
		rows, err = db.conn.QueryContext(ctx, queries.GetFollowing, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to query following: %w", err)
		}
		defer rows.Close()

		var following []FollowerWithDM
		for rows.Next() {
			var f FollowerWithDM
			if scanErr := rows.Scan(
				&f.ID, &f.UUID, &f.Email, &f.FirstName, &f.LastName,
				&f.Nickname, &f.AvatarPath, &f.IsPublic,
			); scanErr != nil {
				return nil, fmt.Errorf("failed to scan following user: %w", scanErr)
			}
			following = append(following, f)
		}

		if scanErr := rows.Err(); scanErr != nil {
			return nil, fmt.Errorf("row iteration error: %w", scanErr)
		}

		return following, nil
	}
	defer rows.Close()

	var following []FollowerWithDM
	for rows.Next() {
		var f FollowerWithDM
		var lastDMAtStr sql.NullString
		if scanErr := rows.Scan(
			&f.ID, &f.UUID, &f.Email, &f.FirstName, &f.LastName,
			&f.Nickname, &f.AvatarPath, &f.IsPublic, &lastDMAtStr, &f.UnreadCount,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan following user: %w", scanErr)
		}
		if lastDMAtStr.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", lastDMAtStr.String); err == nil {
				f.LastDMAt = t
			}
		}
		following = append(following, f)
	}

	if scanErr := rows.Err(); scanErr != nil {
		return nil, fmt.Errorf("row iteration error: %w", scanErr)
	}

	return following, nil
}

// CheckFollowing checks if a user is following another user
func (db *DataBase) CheckFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var exists bool
	err := db.conn.QueryRowContext(ctx,
		queries.CheckFollowing,
		followerID, followingID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check following: %w", err)
	}

	return exists, nil
}

// GetPendingFollowRequests retrieves pending follow requests for a user
func (db *DataBase) GetPendingFollowRequests(ctx context.Context, userID int64) ([]User, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx, queries.GetPendingFollowRequests, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending requests: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var requestID int64
		err := rows.Scan(
			&requestID, &u.ID, &u.UUID, &u.Email, &u.FirstName,
			&u.LastName, &u.Nickname, &u.AvatarPath,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending request: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return users, nil
}

// GetFollowerIDs retrieves just the IDs of followers for a user
func (db *DataBase) GetFollowerIDs(ctx context.Context, userID int64) ([]int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx, queries.GetFollowerIDs, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query follower IDs: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan follower id: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return ids, nil
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
		queries.CreatePost,
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
		queries.GetPostByID,
		postID, userID,
	).Scan(
		&post.ID,
		&post.UUID,
		&post.AuthorID,
		&post.AuthorUUID,
		&post.GroupID,
		&post.GroupUUID,
		&post.Content,
		&post.ImagePath,
		&post.PrivacyLevel,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.IsDeleted,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to query post: %w", err)
	}

	return post, nil
}

// GetPostByUUID retrieves a post by its UUID
func (db *DataBase) GetPostByUUID(ctx context.Context, postUUID string) (*Post, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	post := &Post{}
	err := db.conn.QueryRowContext(ctx,
		queries.GetPostByUUID,
		postUUID,
	).Scan(
		&post.ID,
		&post.UUID,
		&post.AuthorID,
		&post.AuthorUUID,
		&post.GroupID,
		&post.GroupUUID,
		&post.Content,
		&post.ImagePath,
		&post.PrivacyLevel,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.IsDeleted,
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
		queries.GetUserPosts,
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
			&post.AuthorID,
			&post.AuthorUUID,
			&post.GroupID,
			&post.GroupUUID,
			&post.Content,
			&post.ImagePath,
			&post.PrivacyLevel,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.IsDeleted,
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

// GetFeed retrieves the paginated news feed for a user
func (db *DataBase) GetFeed(ctx context.Context, userID int64, limit, offset int) ([]*Post, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx,
		queries.GetFeed,
		userID, userID, userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query feed: %w", err)
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{Author: &User{}}
		var nickname, avatarPath sql.NullString

		err := rows.Scan(
			&post.ID,
			&post.UUID,
			&post.AuthorID,
			&post.AuthorUUID,
			&post.Content,
			&post.ImagePath,
			&post.PrivacyLevel,
			&post.CreatedAt,
			&post.Author.FirstName,
			&post.Author.LastName,
			&nickname,
			&avatarPath,
			&post.IsDeleted,
			&post.CommentCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		if nickname.Valid {
			post.Author.Nickname = &nickname.String
		}
		if avatarPath.Valid {
			post.Author.AvatarPath = &avatarPath.String
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return posts, nil
}

// GetGroupPosts retrieves posts within a group
func (db *DataBase) GetGroupPosts(ctx context.Context, groupID int64, limit, offset int) ([]*Post, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx,
		queries.GetGroupPosts,
		groupID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query group posts: %w", err)
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{Author: &User{}}
		var nickname, avatarPath sql.NullString

		err := rows.Scan(
			&post.ID,
			&post.UUID,
			&post.AuthorID,
			&post.AuthorUUID,
			&post.GroupID,
			&post.GroupUUID,
			&post.Content,
			&post.ImagePath,
			&post.PrivacyLevel,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.IsDeleted,
			&post.Author.FirstName,
			&post.Author.LastName,
			&nickname,
			&avatarPath,
			&post.CommentCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group post: %w", err)
		}
		if nickname.Valid {
			post.Author.Nickname = &nickname.String
		}
		if avatarPath.Valid {
			post.Author.AvatarPath = &avatarPath.String
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return posts, nil
}

// GetPostWithAuthor retrieves a single post with author info
func (db *DataBase) GetPostWithAuthor(ctx context.Context, postID, userID int64) (*Post, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	post := &Post{Author: &User{}}
	var nickname, avatarPath sql.NullString
	err := db.conn.QueryRowContext(ctx, `
		SELECT p.id, p.uuid, p.author_id, u.uuid as author_uuid, p.group_id, g.uuid as group_uuid,
		 p.content, p.image_path, p.privacy_level, p.created_at, p.updated_at,
		 u.first_name, u.last_name, u.nickname, u.avatar_path,
		 p.is_deleted,
		 (SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comment_count
		FROM posts p
		JOIN users u ON u.id = p.author_id
		LEFT JOIN groups g ON g.id = p.group_id
		WHERE p.id = ?
	`, postID).Scan(
		&post.ID,
		&post.UUID,
		&post.AuthorID,
		&post.AuthorUUID,
		&post.GroupID,
		&post.GroupUUID,
		&post.Content,
		&post.ImagePath,
		&post.PrivacyLevel,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Author.FirstName,
		&post.Author.LastName,
		&nickname,
		&avatarPath,
		&post.IsDeleted,
		&post.CommentCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to query post: %w", err)
	}
	if nickname.Valid {
		post.Author.Nickname = &nickname.String
	}
	if avatarPath.Valid {
		post.Author.AvatarPath = &avatarPath.String
	}

	return post, nil
}

// DeletePost soft deletes a post
func (db *DataBase) DeletePost(ctx context.Context, postID, userID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		queries.DeletePost,
		postID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}

// UpdatePost updates a post's content and privacy
func (db *DataBase) UpdatePost(ctx context.Context, postID, authorID int64, content string, imagePath *string, privacyLevel string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		queries.UpdatePost,
		content, imagePath, privacyLevel, postID, authorID,
	)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("post not found or not authorized")
	}

	return nil
}

//====================================
// COMMENT METHODS
//====================================

// CreateComment creates a new comment
func (db *DataBase) CreateComment(ctx context.Context, comment *Comment) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		queries.CreateComment,
		comment.UUID,
		comment.PostID,
		comment.AuthorID,
		comment.ParentCommentID,
		comment.Content,
		comment.ImagePath,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create comment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get comment id: %w", err)
	}

	return id, nil
}

// GetPostComments retrieves all comments for a post (flat list)
func (db *DataBase) GetPostComments(ctx context.Context, postID int64) ([]*Comment, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx,
		queries.GetPostComments,
		postID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		c := &Comment{}
		var firstName, lastName string
		var nickname, avatarPath sql.NullString

		err := rows.Scan(
			&c.ID,
			&c.UUID,
			&c.PostID,
			&c.PostUUID,
			&c.AuthorID,
			&c.AuthorUUID,
			&c.ParentCommentID,
			&c.ParentCommentUUID,
			&c.Content,
			&c.ImagePath,
			&c.CreatedAt,
			&c.UpdatedAt,
			&firstName,
			&lastName,
			&nickname,
			&avatarPath,
			&c.IsDeleted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		c.Author = &User{
			FirstName: firstName,
			LastName:  lastName,
		}
		if nickname.Valid {
			c.Author.Nickname = &nickname.String
		}
		if avatarPath.Valid {
			c.Author.AvatarPath = &avatarPath.String
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return comments, nil
}

// DeleteComment soft-deletes a comment (preserves the post id)
func (db *DataBase) DeleteComment(ctx context.Context, commentID, authorID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		queries.DeleteComment,
		commentID, authorID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}

// UpdateComment updates a comment's content
func (db *DataBase) UpdateComment(ctx context.Context, commentID, authorID int64, content string, imagePath *string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		queries.UpdateComment,
		content, imagePath, commentID, authorID,
	)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("comment not found or not authorized")
	}

	return nil
}

// GetComment retrieves a single comment by ID
func (db *DataBase) GetComment(ctx context.Context, commentID int64) (*Comment, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	c := &Comment{}
	err := db.conn.QueryRowContext(ctx,
		queries.GetCommentByID,
		commentID,
	).Scan(
		&c.ID,
		&c.UUID,
		&c.PostID,
		&c.PostUUID,
		&c.AuthorID,
		&c.AuthorUUID,
		&c.ParentCommentID,
		&c.ParentCommentUUID,
		&c.Content,
		&c.ImagePath,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to query comment: %w", err)
	}

	return c, nil
}

// GetCommentByUUID retrieves a single comment by its UUID
func (db *DataBase) GetCommentByUUID(ctx context.Context, commentUUID string) (*Comment, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	c := &Comment{}
	err := db.conn.QueryRowContext(ctx,
		queries.GetCommentByUUID,
		commentUUID,
	).Scan(
		&c.ID,
		&c.UUID,
		&c.PostID,
		&c.PostUUID,
		&c.AuthorID,
		&c.AuthorUUID,
		&c.ParentCommentID,
		&c.ParentCommentUUID,
		&c.Content,
		&c.ImagePath,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.IsDeleted,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to query comment: %w", err)
	}

	return c, nil
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
		queries.CreateGroup,
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
		queries.GetGroupByID,
		groupID,
	).Scan(
		&group.ID,
		&group.UUID,
		&group.CreatorID,
		&group.CreatorUUID,
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

// GetGroupByUUID retrieves a group by its UUID
func (db *DataBase) GetGroupByUUID(ctx context.Context, uuid string) (*Group, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	group := &Group{}
	err := db.conn.QueryRowContext(ctx,
		queries.GetGroupByUUID,
		uuid,
	).Scan(
		&group.ID,
		&group.UUID,
		&group.CreatorID,
		&group.CreatorUUID,
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
		queries.GetUserGroups,
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

// UpdateGroup updates a group's details
func (db *DataBase) UpdateGroup(ctx context.Context, groupID int64, title, description string, avatarPath *string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		`UPDATE groups SET title = ?, description = ?, avatar_path = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		title, description, avatarPath, groupID,
	)
	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	return nil
}

// UpdateGroupAvatar updates a group's avatar path
func (db *DataBase) UpdateGroupAvatar(ctx context.Context, groupID int64, avatarPath string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		queries.UpdateGroupAvatar,
		avatarPath, groupID,
	)
	if err != nil {
		return fmt.Errorf("failed to update group avatar: %w", err)
	}

	return nil
}

// AddGroupMember adds a user to a group
func (db *DataBase) AddGroupMember(ctx context.Context, groupID, userID int64, invitedBy *int64, status string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		queries.AddGroupMember,
		groupID, userID, status, invitedBy, invitedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to add group member: %w", err)
	}

	return nil
}

// UpdateMemberStatus updates a group member's status
func (db *DataBase) UpdateMemberStatus(ctx context.Context, groupID, userID int64, status string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		queries.UpdateMemberStatus,
		status, status, groupID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update member status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

// GetGroupMembers retrieves members of a group
func (db *DataBase) GetGroupMembers(ctx context.Context, groupID int64) ([]GroupMember, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx, queries.GetGroupMembers, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query group members: %w", err)
	}
	defer rows.Close()

	var members []GroupMember
	for rows.Next() {
		var gm GroupMember
		u := &User{}
		var nickname, avatarPath, invitedByUUID sql.NullString

		err := rows.Scan(
			&u.ID, &u.UUID, &u.Email, &u.FirstName, &u.LastName,
			&nickname, &avatarPath, &gm.Status, &gm.JoinedAt,
			&invitedByUUID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group member: %w", err)
		}
		if nickname.Valid {
			u.Nickname = &nickname.String
		}
		if avatarPath.Valid {
			u.AvatarPath = &avatarPath.String
		}
		gm.GroupID = groupID
		gm.UserID = u.ID
		gm.User = u
		if invitedByUUID.Valid {
			gm.InvitedByUUID = &invitedByUUID.String
		}
		members = append(members, gm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return members, nil
}

// GetAllGroups retrieves all groups with pagination for browsing
func (db *DataBase) GetAllGroups(ctx context.Context, limit, offset int) ([]Group, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx, queries.GetAllGroups, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query all groups: %w", err)
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		var memberCount int
		err := rows.Scan(
			&g.ID, &g.UUID, &g.CreatorID, &g.Title, &g.Description,
			&g.AvatarPath, &g.CreatedAt, &memberCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, g)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return groups, nil
}

// DeleteGroup removes a group (creator only)
func (db *DataBase) DeleteGroup(ctx context.Context, groupID, creatorID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		`DELETE FROM groups WHERE id = ? AND creator_id = ?`,
		groupID, creatorID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("group not found or not authorized")
	}

	return nil
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
		queries.CreateMessage,
		message.UUID,
		message.SenderID,
		message.DirectMessageID,
		message.GroupID,
		message.Content,
		message.ImagePath,
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

	// Normalize user IDs so that user1 is always the smaller ID
	u1, u2 := dm.User1ID, dm.User2ID
	if u1 > u2 {
		u1, u2 = u2, u1
	}

	// Try to find or create with normalized order
	var dmID int64
	err := db.conn.QueryRowContext(dbCtx,
		queries.GetOrCreateDM,
		u1, u2,
	).Scan(&dmID)
	if err != nil {
		// If the normalized insert fails (e.g., old reversed-order DM exists),
		// try the reverse order which matches the old record
		err = db.conn.QueryRowContext(dbCtx,
			queries.GetDMidScalar,
			u1, u2, u2, u1,
		).Scan(&dmID)
		if err != nil {
			return 0, fmt.Errorf("failed to create or get direct message: %w", err)
		}
	}

	return dmID, nil
}

// GetUnreadMessageCount returns total unread DM message count for a user
func (db *DataBase) GetUnreadMessageCount(ctx context.Context, userID int64) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var count int
	err := db.conn.QueryRowContext(ctx, queries.GetUnreadDMCount, userID, userID, userID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread messages: %w", err)
	}

	return count, nil
}

// GetUnreadCountForDM returns unread message count for a specific DM
func (db *DataBase) GetUnreadCountForDM(ctx context.Context, dmID, userID int64) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var count int
	err := db.conn.QueryRowContext(ctx, queries.GetUnreadCountForDM, dmID, userID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread for dm: %w", err)
	}

	return count, nil
}

// GetAllDMs retrieves all DM conversations for a user with the other user's info
func (db *DataBase) GetAllDMs(ctx context.Context, userID int64) ([]*DirectMessage, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx, queries.GetAllDMs, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query DMs: %w", err)
	}
	defer rows.Close()

	var dms []*DirectMessage
	for rows.Next() {
		var dm DirectMessage
		var otherUserID int64
		var dmLastMessageAt sql.NullTime
		err := rows.Scan(
			&dm.ID,
			&otherUserID,
			&dm.CreatedAt,
			&dmLastMessageAt,
			&dm.LastMessage,
			&dm.LastMessageAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan DM: %w", err)
		}

		otherUser, err := db.GetUserByID(ctx, otherUserID)
		if err == nil {
			dm.OtherUser = otherUser
		}

		if dm.OtherUser != nil {
			dm.OtherUser.IsOnline = false
		}

		dms = append(dms, &dm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return dms, nil
}

// GetMessages retrieves messages for a direct message conversation
func (db *DataBase) GetMessages(ctx context.Context, dmID int64, limit int) ([]*Message, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx, queries.GetPrivateMessages, dmID, nil, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		var senderUser User
		var firstName, lastName string
		var nickname, avatarPath, imagePath *string
		err := rows.Scan(
			&msg.ID, &msg.UUID, &msg.SenderID, &msg.Content,
			&imagePath, &msg.IsRead, &msg.CreatedAt,
			&firstName, &lastName, &nickname, &avatarPath, &senderUser.UUID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		msg.ImagePath = imagePath
		senderUser.FirstName = firstName
		senderUser.LastName = lastName
		senderUser.Nickname = nickname
		senderUser.AvatarPath = avatarPath
		senderUser.ID = msg.SenderID
		msg.Sender = &senderUser
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return messages, nil
}

// GetDMByUsers finds the DM between two users
func (db *DataBase) GetDMByUsers(ctx context.Context, user1ID, user2ID int64) (*DirectMessage, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var dm DirectMessage
	err := db.conn.QueryRowContext(ctx, queries.GetDMid, user1ID, user2ID, user2ID, user1ID).Scan(
		&dm.ID, &dm.User1ID, &dm.User2ID, &dm.CreatedAt, &dm.LastMessageAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get DM: %w", err)
	}

	return &dm, nil
}

// MarkMessageRead marks a message as read by a user
func (db *DataBase) MarkMessageRead(ctx context.Context, messageID, userID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx, queries.MarkMessageRead, messageID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark message read: %w", err)
	}

	return nil
}

// UpdateDMTime updates the last message time for a direct message
func (db *DataBase) UpdateDMTime(ctx context.Context, dmID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx, queries.UpdateLastDM, dmID)
	if err != nil {
		return fmt.Errorf("failed to update DM time: %w", err)
	}

	return nil
}

// GetFollowersWithDM retrieves followers sorted by last DM time then alphabetically
func (db *DataBase) GetFollowersWithDM(ctx context.Context, userID int64) ([]FollowerWithDM, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx, queries.GetFollowersWithDM, userID, userID, userID)
	if err != nil {
		rows, err = db.conn.QueryContext(ctx, queries.GetFollowers, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to query followers: %w", err)
		}
		defer rows.Close()

		var followers []FollowerWithDM
		for rows.Next() {
			var f FollowerWithDM
			if scanErr := rows.Scan(
				&f.ID, &f.UUID, &f.Email, &f.FirstName, &f.LastName,
				&f.Nickname, &f.AvatarPath, &f.IsPublic,
			); scanErr != nil {
				return nil, fmt.Errorf("failed to scan follower: %w", scanErr)
			}
			followers = append(followers, f)
		}

		if scanErr := rows.Err(); scanErr != nil {
			return nil, fmt.Errorf("row iteration error: %w", scanErr)
		}

		return followers, nil
	}
	defer rows.Close()

	var followers []FollowerWithDM
	for rows.Next() {
		var f FollowerWithDM
		var lastDMAtStr sql.NullString
		if scanErr := rows.Scan(
			&f.ID, &f.UUID, &f.Email, &f.FirstName, &f.LastName,
			&f.Nickname, &f.AvatarPath, &f.IsPublic, &lastDMAtStr, &f.UnreadCount,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan follower: %w", scanErr)
		}
		if lastDMAtStr.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", lastDMAtStr.String); err == nil {
				f.LastDMAt = t
			}
		}
		followers = append(followers, f)
	}

	if scanErr := rows.Err(); scanErr != nil {
		return nil, fmt.Errorf("row iteration error: %w", scanErr)
	}

	return followers, nil
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
		queries.CreateNotification,
		notification.UserID,
		notification.FromUserID,
		notification.Type,
		notification.Content,
		notification.RelatedID,
		notification.RelatedUUID,
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
		queries.GetUserNotifications,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		notif := &Notification{}
		var fromUserUUID, firstName, lastName sql.NullString
		var nickname, avatarPath sql.NullString

		err := rows.Scan(
			&notif.ID,
			&notif.FromUserID,
			&fromUserUUID,
			&firstName,
			&lastName,
			&nickname,
			&avatarPath,
			&notif.Type,
			&notif.Content,
			&notif.IsRead,
			&notif.RelatedID,
			&notif.RelatedUUID,
			&notif.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if fromUserUUID.Valid {
			notif.FromUserUUID = &fromUserUUID.String
			fromUser := &User{
				UUID:      fromUserUUID.String,
				FirstName: firstName.String,
				LastName:  lastName.String,
			}
			if nickname.Valid {
				fromUser.Nickname = &nickname.String
			}
			if avatarPath.Valid {
				fromUser.AvatarPath = &avatarPath.String
			}
			notif.FromUser = fromUser
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
		queries.MarkNotificationsAsRead,
		notificationID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

// UpdateGroupJoinNotification updates a user's join-request notification for a
// group with the accept/decline outcome, marking it read so the bell/notifications
// stay in sync after accept/decline.
func (db *DataBase) UpdateGroupJoinNotification(ctx context.Context, groupUUID string, requesterID int64, content string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		queries.UpdateGroupJoinNotification,
		content, groupUUID, requesterID,
	)
	if err != nil {
		return fmt.Errorf("failed to update group join notification: %w", err)
	}

	return nil
}

// MarkAllNotificationsAsRead marks all notifications for a user as read
func (db *DataBase) MarkAllNotificationsAsRead(ctx context.Context, userID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		`UPDATE notifications SET is_read = 1, read_at = CURRENT_TIMESTAMP WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// GetUnreadNotificationCount returns the number of unread notifications for a user
func (db *DataBase) GetUnreadNotificationCount(ctx context.Context, userID int64) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var count int
	err := db.conn.QueryRowContext(ctx, queries.GetUnreadCount, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

//====================================
// EVENT METHODS
//====================================

// CreateEvent creates a new group event
func (db *DataBase) CreateEvent(ctx context.Context, event *Event) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := db.conn.ExecContext(dbCtx,
		queries.CreateEvent,
		event.UUID,
		event.GroupID,
		event.CreatorID,
		event.Title,
		event.Description,
		event.ImagePath,
		event.EventDateTime,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create event: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get event id: %w", err)
	}

	return id, nil
}

// GetEventByID retrieves an event by ID with the creator's info
func (db *DataBase) GetEventByID(ctx context.Context, eventID int64) (*Event, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	event := &Event{}
	var firstName, lastName string
	var nickname sql.NullString
	var imagePath *string

	err := db.conn.QueryRowContext(ctx,
		queries.GetEventByID,
		eventID,
	).Scan(
		&event.ID,
		&event.UUID,
		&event.GroupID,
		&event.CreatorID,
		&event.Title,
		&event.Description,
		&imagePath,
		&event.EventDateTime,
		&event.CreatedAt,
		&event.UpdatedAt,
		&firstName,
		&lastName,
		&nickname,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to query event: %w", err)
	}

	event.ImagePath = imagePath
	event.Creator = &User{
		FirstName: firstName,
		LastName:  lastName,
	}
	if nickname.Valid {
		event.Creator.Nickname = &nickname.String
	}

	return event, nil
}

// GetEventByUUID retrieves an event by its UUID with the creator's info
func (db *DataBase) GetEventByUUID(ctx context.Context, uuid string) (*Event, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	event := &Event{}
	var firstName, lastName string
	var nickname sql.NullString
	var imagePath *string

	err := db.conn.QueryRowContext(ctx,
		queries.GetEventByUUID,
		uuid,
	).Scan(
		&event.ID,
		&event.UUID,
		&event.GroupID,
		&event.CreatorID,
		&event.Title,
		&event.Description,
		&imagePath,
		&event.EventDateTime,
		&event.CreatedAt,
		&event.UpdatedAt,
		&firstName,
		&lastName,
		&nickname,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to query event: %w", err)
	}

	event.ImagePath = imagePath
	event.Creator = &User{
		FirstName: firstName,
		LastName:  lastName,
	}
	if nickname.Valid {
		event.Creator.Nickname = &nickname.String
	}

	return event, nil
}

// GetGroupEvents retrieves all events for a group
func (db *DataBase) GetGroupEvents(ctx context.Context, groupID int64) ([]*Event, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx,
		queries.GetGroupEvents,
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query group events: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		event := &Event{}
		var firstName, lastName string
		var nickname sql.NullString
		var imagePath *string

		err := rows.Scan(
			&event.ID,
			&event.UUID,
			&event.GroupID,
			&event.CreatorID,
			&event.Title,
			&event.Description,
			&imagePath,
			&event.EventDateTime,
			&event.CreatedAt,
			&event.UpdatedAt,
			&firstName,
			&lastName,
			&nickname,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		event.ImagePath = imagePath

		event.Creator = &User{
			FirstName: firstName,
			LastName:  lastName,
		}
		if nickname.Valid {
			event.Creator.Nickname = &nickname.String
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return events, nil
}

// CreateEventRSVP upserts a user's response to an event
func (db *DataBase) CreateEventRSVP(ctx context.Context, eventID, userID int64, response string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		queries.CreateEventRSVP,
		eventID, userID, response, response,
	)
	if err != nil {
		return fmt.Errorf("failed to save event response: %w", err)
	}

	return nil
}

// GetEventResponseByUser retrieves a user's response to an event
func (db *DataBase) GetEventResponseByUser(ctx context.Context, eventID, userID int64) (*EventResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	resp := &EventResponse{}
	err := db.conn.QueryRowContext(ctx,
		queries.GetEventResponseByUser,
		eventID, userID,
	).Scan(
		&resp.ID,
		&resp.EventID,
		&resp.UserID,
		&resp.Response,
		&resp.CreatedAt,
		&resp.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query event response: %w", err)
	}

	return resp, nil
}

// GetEventResponseCounts returns the number of going / not_going responses for an event
func (db *DataBase) GetEventResponseCounts(ctx context.Context, eventID int64) (map[string]int, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := db.conn.QueryContext(ctx,
		queries.GetEventResponseCounts,
		eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query event response counts: %w", err)
	}
	defer rows.Close()

	counts := map[string]int{"going": 0, "not_going": 0}
	for rows.Next() {
		var response string
		var count int
		if err := rows.Scan(&response, &count); err != nil {
			return nil, fmt.Errorf("failed to scan response count: %w", err)
		}
		counts[response] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return counts, nil
}

// CreateOAuthAccount links a user to an OAuth provider account.
func (db *DataBase) CreateOAuthAccount(ctx context.Context, userID int64, provider, providerID string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(dbCtx,
		queries.CreateOAuthAccount,
		userID,
		provider,
		providerID,
	)
	if err != nil {
		return fmt.Errorf("failed to create oauth account: %w", err)
	}

	return nil
}

// GetUserIDByOAuthAccount returns the user ID linked to a provider account, or
// sql.ErrNoRows if no such link exists.
func (db *DataBase) GetUserIDByOAuthAccount(ctx context.Context, provider, providerID string) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var userID int64
	err := db.conn.QueryRowContext(ctx,
		queries.GetOAuthAccount,
		provider,
		providerID,
	).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, sql.ErrNoRows
		}
		return 0, fmt.Errorf("failed to query oauth account: %w", err)
	}

	return userID, nil
}
