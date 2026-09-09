package database

import (
	"database/sql"
	"time"
)

//====================================
// MODELS
//====================================

type User struct {
	ID           int64     `json:"-"`
	UUID         string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Nickname     *string   `json:"nickname,omitempty"`
	DateOfBirth  time.Time `json:"date_of_birth"`
	AboutMe      *string   `json:"about_me,omitempty"`
	AvatarPath   *string   `json:"avatar_path,omitempty"`
	IsPublic     bool      `json:"is_public"`
	IsActive     bool      `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastSeen     time.Time `json:"last_seen,omitempty"`
	IsOnline     bool      `json:"is_online,omitempty"`
}

// UserSearchResult is a user with the viewer's follow state attached, used for
// the user search endpoint.
type UserSearchResult struct {
	ID              int64
	UUID            string
	FirstName       string
	LastName        string
	Nickname        *string
	AvatarPath      *string
	IsPublic        bool
	IsFollowing     bool
	IsFollowPending bool
}

type UserQueries struct {
	Create        *sql.Stmt
	GetByEmail    *sql.Stmt
	GetByID       *sql.Stmt
	GetByUUID     *sql.Stmt
	UpdatePrivacy *sql.Stmt
	UpdateProfile *sql.Stmt
}

type Session struct {
	ID        int64  `json:"-"`
	SessionID string `json:"session_id"`
	UserID    int64  `json:"user_id"`
	IPAddress string
	UserAgent string
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool
	CreatedAt time.Time
}

type SessionQueries struct {
	Create           *sql.Stmt
	Get              *sql.Stmt
	Delete           *sql.Stmt
	Cleanup          *sql.Stmt
	GetUserBySession *sql.Stmt
}

type Follow struct {
	FollowerID  int64     `json:"follower_id"`
	FollowingID int64     `json:"following_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type FollowQueries struct {
	Create       *sql.Stmt
	Update       *sql.Stmt
	GetRequest   *sql.Stmt
	GetFollowers *sql.Stmt
	GetFollowing *sql.Stmt
	Check        *sql.Stmt
	GetPending   *sql.Stmt
}

type FollowerWithDM struct {
	User
	UnreadCount int       `json:"unread_count"`
	LastDMAt    time.Time `json:"last_dm_at,omitempty"`
}

type Post struct {
	ID           int64     `json:"-"`
	UUID         string    `json:"id"`
	AuthorID     int64     `json:"-"`
	AuthorUUID   string    `json:"author_id"`
	Author       *User     `json:"author,omitempty"`
	GroupID      *int64    `json:"-"`
	GroupUUID    *string   `json:"group_id,omitempty"`
	Content      string    `json:"content"`
	ImagePath    *string   `json:"image_path,omitempty"`
	PrivacyLevel string    `json:"privacy_level"`
	IsDeleted    bool      `json:"is_deleted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Comments     []Comment `json:"comments,omitempty"`
	CommentCount int       `json:"comment_count,omitempty"`
}

type Comment struct {
	ID                int64     `json:"-"`
	UUID              string    `json:"id"`
	PostID            int64     `json:"-"`
	PostUUID          string    `json:"post_id"`
	AuthorID          int64     `json:"-"`
	AuthorUUID        string    `json:"author_id"`
	Author            *User     `json:"author,omitempty"`
	ParentCommentID   *int64    `json:"-"`
	ParentCommentUUID *string   `json:"parent_comment_id,omitempty"`
	Content           string    `json:"content"`
	ImagePath         *string   `json:"image_path,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	IsDeleted         bool      `json:"is_deleted"`
}

type CommentQueries struct {
	Create          *sql.Stmt
	GetPostComments *sql.Stmt
	Delete          *sql.Stmt
}

type PostQueries struct {
	Create           *sql.Stmt
	GetByID          *sql.Stmt
	GetFeed          *sql.Stmt
	Delete           *sql.Stmt
	AddVisibility    *sql.Stmt
	GetVisible       *sql.Stmt
	RemoveVisibility *sql.Stmt
	GetUserPosts     *sql.Stmt
}

type Group struct {
	ID            int64      `json:"-"`
	UUID          string     `json:"id"`
	CreatorID     int64      `json:"-"`
	CreatorUUID   string     `json:"creator_id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	AvatarPath    *string    `json:"avatar_path,omitempty"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type GroupMember struct {
	ID            int64     `json:"-"`
	GroupID       int64     `json:"-"`
	UserID        int64     `json:"-"`
	User          *User     `json:"user,omitempty"`
	Status        string    `json:"status"`
	InvitedBy     int64     `json:"-"`
	InvitedByUUID *string   `json:"invited_by,omitempty"`
	JoinedAt      time.Time `json:"joined_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type GroupQueries struct {
	Create        *sql.Stmt
	GetByID       *sql.Stmt
	GetUserGroups *sql.Stmt
	AddMember     *sql.Stmt
	UpdateStatus  *sql.Stmt
	GetMembers    *sql.Stmt
	GetAllGroups  *sql.Stmt
}

type Event struct {
	ID            int64     `json:"-"`
	UUID          string    `json:"id"`
	GroupID       int64     `json:"group_id"`
	CreatorID     int64     `json:"creator_id"`
	Creator       *User     `json:"creator,omitempty"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	ImagePath     *string   `json:"image_path,omitempty"`
	EventDateTime time.Time `json:"event_datetime"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type EventResponse struct {
	ID        int64     `json:"-"`
	EventID   int64     `json:"event_id"`
	UserID    int64     `json:"user_id"`
	Response  string    `json:"response"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type EventQueries struct {
	Create         *sql.Stmt
	RSVP           *sql.Stmt
	GetGroupEvents *sql.Stmt
}

type Message struct {
	ID              int64     `json:"-"`
	UUID            string    `json:"uuid"`
	SenderID        int64     `json:"sender_id"`
	Sender          *User     `json:"sender,omitempty"`
	DirectMessageID *int64    `json:"direct_message_id,omitempty"`
	GroupID         *int64    `json:"group_id,omitempty"`
	Content         string    `json:"content"`
	ImagePath       *string   `json:"image_path,omitempty"`
	IsRead          bool      `json:"is_read"`
	CreatedAt       time.Time `json:"created_at"`
}

type DirectMessage struct {
	ID            int64      `json:"id"`
	User1ID       int64      `json:"-"`
	User2ID       int64      `json:"-"`
	OtherUser     *User      `json:"other_user"`
	LastMessage   *string    `json:"last_message,omitempty"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type MessageQueries struct {
	Create        *sql.Stmt
	GetByID       *sql.Stmt
	GetOrCreate   *sql.Stmt
	GetGroup      *sql.Stmt
	GetAll        *sql.Stmt
	GetPrivate    *sql.Stmt
	MarkRead      *sql.Stmt
	UpdateMessage *sql.Stmt
	UpdateDM      *sql.Stmt
}

type Notification struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"-"`
	FromUserID   *int64    `json:"-"`
	FromUserUUID *string   `json:"from_user_id,omitempty"`
	FromUser     *User     `json:"from_user,omitempty"`
	Type         string    `json:"type"`
	Content      string    `json:"content"`
	RelatedID    *int64    `json:"-"`
	RelatedUUID  *string   `json:"related_id,omitempty"`
	IsRead       bool      `json:"is_read"`
	ReadAt       time.Time `json:"read_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type NotificationQueries struct {
	Create    *sql.Stmt
	Get       *sql.Stmt
	MarkRead  *sql.Stmt
	GetUnread *sql.Stmt
}
