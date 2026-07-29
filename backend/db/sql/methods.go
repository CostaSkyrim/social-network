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
		&post.GroupID,
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

// GetPostWithAuthor retrieves a single post with author info
func (db *DataBase) GetPostWithAuthor(ctx context.Context, postID, userID int64) (*Post, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	post := &Post{Author: &User{}}
	var nickname, avatarPath sql.NullString
	err := db.conn.QueryRowContext(ctx, `
		SELECT p.id, p.uuid, p.author_id, p.content, p.image_path,
		 p.privacy_level, p.created_at, p.updated_at,
		 u.first_name, u.last_name, u.nickname, u.avatar_path,
		 p.is_deleted,
		 (SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comment_count
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.id = ?
	`, postID).Scan(
		&post.ID,
		&post.UUID,
		&post.AuthorID,
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
			&c.AuthorID,
			&c.ParentCommentID,
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
		&c.AuthorID,
		&c.ParentCommentID,
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

// AddGroupMember adds a user to a group
func (db *DataBase) AddGroupMember(ctx context.Context, groupID, userID, invitedBy int64, status string) error {
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
		var u struct {
			ID         int64
			UUID       string
			Email      string
			FirstName  string
			LastName   string
			Nickname   *string
			AvatarPath *string
		}
		err := rows.Scan(
			&u.ID, &u.UUID, &u.Email, &u.FirstName, &u.LastName,
			&u.Nickname, &u.AvatarPath, &gm.Status, &gm.JoinedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group member: %w", err)
		}
		gm.GroupID = groupID
		gm.UserID = u.ID
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
		queries.GetOrCreateDM,
		dm.User1ID, dm.User2ID, dm.User2ID, dm.User1ID,
	).Scan(&dmID)

	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query direct message: %w", err)
	}

	if err == sql.ErrNoRows {
		// Create new DM
		result, err := db.conn.ExecContext(dbCtx,
			queries.GetOrCreateDM,
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
		queries.CreateNotification,
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
		queries.MarkNotificationsAsRead,
		notificationID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}
