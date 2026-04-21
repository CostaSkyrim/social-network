package queries

// User queries
const (
	CreateUser = `
		INSERT INTO users (uuid, email, password_hash,
		 first_name, last_name, nickname, date_of_birth, 
		 about_me, avatar_path, is_public)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	GetUserByEmail = `
		SELECT id, uuid, email, password_hash, first_name,
		 last_name, nickname, date_of_birth, about_me, avatar_path,
		 is_public, is_active, created_at, updated_at
		FROM users
		WHERE email = ? AND is_active = 1
	`

	GetUserByID = `
		SELECT id, uuid, email, first_name, last_name, nickname,
		 date_of_birth, about_me, avatar_path, is_public, created_at
		FROM users
		WHERE id = ? AND is_active = 1
	`

	GetUserByUUID = `
		SELECT id, uuid, email, first_name, last_name, nickname,
		 date_of_birth, about_me, avatar_path, is_public, created_at
		FROM users
		WHERE uuid = ? AND is_active = 1
	`

	UpdateUserPrivacy = `
		UPDATE users
		SET is_public = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	UpdateUserProfile = `
		UPDATE users
		SET nickname = COALESCE(?, nickname),
			about_me = COALESCE(?, about_me),
			avatar_path = COALESCE(?, avatar_path),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
)

// Session queries
const (
	CreateSession = `
		INSERT INTO sessions (session_id, user_id, ip_address, user_agent,
		 expires_at)
		VALUES (?, ?, ?, ?, ?)
	`

	GetSession = `
		SELECT id, session_id, user_id, expires_at, is_active
		FROM sessions
		WHERE session_id = ? AND expires_at > CURRENT_TIMESTAMP AND is_active = 1
	`

	DeleteSession = `
		UPDATE sessions
		SET is_active = 0
		WHERE session_id = ?
	`

	CleanupSessions = `
		UPDATE sessions
		SET is_active = 0
		WHERE expires_at < CURRENT_TIMESTAMP
	`
)

// Follow queries
const (
	CreateFollowRequest = `
		INSERT INTO followers (follower_id, following_id, status)
		VALUES (?, ?, 'pending')
		ON CONFLICT(follower_id, following_id) DO UPDATE SET
			status = 'pending',
			updated_at = CURRENT_TIMESTAMP
	`

	UpdateFollowStatus = `
		UPDATE followers
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE follower_id = ? AND following_id = ?
	`

	GetFollowRequest = `
		SELECT id, status
		FROM followers
		WHERE follower_id = ? AND following_id = ?
	`

	GetFollowers = `
		SELECT u.id, u.uuid, u.email, u.first_name, u.last_name,
		 u.nickname, u.avatar_path, u.is_public
		FROM followers f
		JOIN users u ON u.id = f.follower_id
		WHERE f.following_id = ? AND f.status = 'accepted' AND u.is_active = 1
		ORDER BY f.created_at DESC
	`

	GetFollowing = `
		SELECT u.id, u.uuid, u.email, u.first_name, u.last_name,
		 u.nickname, u.avatar_path, u.is_public
		FROM followers f
		JOIN users u ON u.id = f.following_id
		WHERE f.follower_id = ? AND f.status = 'accepted' AND u.is_active = 1
		ORDER BY f.created_at DESC
	`

	CheckFollowing = `
		SELECT EXISTS(
			SELECT 1 FROM followers
			WHERE follower_id = ? AND following_id = ? AND status = 'accepted'
		)
	`

	GetPendingFollowRequests = `
		SELECT f.id, u.id, u.uuid, u.email, u.first_name, u.last_name,
		 u.nickname, u.avatar_path
		FROM followers f
		JOIN users u ON u.id = f.follower_id
		WHERE f.following_id = ? AND f.status = 'pending'
		ORDER BY f.created_at DESC
	`
)

// Post queries
const (
	CreatePost = `
		INSERT INTO posts (uuid, author_id, group_id, content, image_path,
		 privacy_level)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	GetPostByID = `
		SELECT id, uuid, author_id, group_id, content, image_path,
		 privacy_level, created_at, updated_at
		FROM posts
		WHERE id = ? AND (group_id IS NULL OR group_id IN (SELECT group_id 
			FROM group_members WHERE user_id = ? AND status = 'accepted'))
	`

	GetUserPosts = `
		SELECT id, uuid, content, image_path, privacy_level, created_at,
		 updated_at
		FROM posts
		WHERE author_id = ? AND group_id IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	DeletePost = `
		UPDATE posts
		SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND author_id = ?
	`
)

// Group queries
const (
	CreateGroup = `
		INSERT INTO groups (uuid, creator_id, title, description,
		 avatar_path)
		VALUES (?, ?, ?, ?, ?)
	`

	GetGroupByID = `
		SELECT id, uuid, creator_id, title, description, avatar_path,
		 created_at, updated_at, last_message_at
		FROM groups
		WHERE id = ?
	`

	GetUserGroups = `
		SELECT g.id, g.uuid, g.title, g.description, g.avatar_path,
		 g.last_message_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.user_id = ? AND gm.status = 'accepted'
		ORDER BY g.last_message_at DESC NULLS LAST
	`

	AddGroupMember = `
		INSERT INTO group_members (group_id, user_id, status, invited_by,
		 joined_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(group_id, user_id) DO UPDATE SET
			status = CASE
				WHEN status = 'declined' THEN 'pending'
				ELSE status
			END,
			invited_by = ?,
			updated_at = CURRENT_TIMESTAMP
	`

	UpdateMemberStatus = `
		UPDATE group_members
		SET status = ?, joined_at = CASE WHEN ? = 'accepted' THEN CURRENT_TIMESTAMP
		 ELSE joined_at END,
		 	updated_at = CURRENT_TIMESTAMP
		WHERE group_id = ? AND user_id = ?
	`

	GetGroupMembers = `
		SELECT u.id, u.uuid, u.email, u.first_name, u.last_name, u.nickname,
		 u.avatar_path, gm.status, gm.joined_at
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = ? AND u.is_active = 1
		ORDER BY gm.joined_at DESC
	`
)

// Message queries
const (
	GetOrCreateDM = `
		INSERT INTO direct_messages (user1_id, user2_id)
		SELECT ?, ? 
		WHERE NOT EXISTS (
			SELECT 1 FROM direct_messages
			WHERE (user1_id = ? AND user2_id = ?) 
			OR (user1_id = ? AND user2_id = ?)
		)
	`
	GetDMid = `
		SELECT id, user1_id, user2_id, created_at, last_message_at
		FROM direct_messages
		WHERE (user1_id = ? AND user2_id = ?) 
		OR (user1_id = ? AND user2_id = ?)
	`

	CreateMessage = `
		INSERT INTO messages (uuid, sender_id, direct_message_id, group_id, content)
		VALUES (?, ?, ?, ?, ?)
	`

	GetGroupMessages = `
		SELECT m.id, m.uuid, m.sender_id, m.content, m.is_read,
		 m.created_at, u.first_name, u.last_name, u.nickname, u.avatar_path
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.group_id = ? AND m.created_at > COALESCE(?, '1970-01-01')
		ORDER BY m.created_at ASC
		LIMIT ?
	`

	GetAllDMs = `
		SELECT DISTINCT
			dm.id, 
			CASE 
				WHEN dm.user1_id = ? THEN dm.user2_id
				ELSE dm.user1_id
			END as other_user_id,
			dm.created_at,
			dm.last_message_at,
			(
				SELECT content FROM messages 
				WHERE direct_message_id = dm.id 
				ORDER BY created_at DESC 
				LIMIT 1
			) as last_message,
			(
				SELECT created_at FROM messages 
				WHERE direct_message_id = dm.id 
				ORDER BY created_at DESC 
				LIMIT 1
			) as last_message_time
		FROM direct_messages dm
		WHERE dm.user1_id = ? OR dm.user2_id = ?
		ORDER BY dm.last_message_at DESC NULLS LAST
	`

	GetPrivateMessages = `
		SELECT m.id, m.uuid, m.sender_id, m.content, m.is_read, m.created_at,
			u.first_name, u.last_name, u.nickname, u.avatar_path
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.direct_message_id = ? 
		AND m.created_at > COALESCE(?, '1970-01-01')
		ORDER BY m.created_at ASC
		LIMIT ?
	`

	MarkMessageRead = `
		INSERT INTO message_reads (message_id, user_id)
		VALUES (?, ?)
		ON CONFLICT(message_id, user_id) DO NOTHING
	`

	UpdateLastMessage = `
		UPDATE groups
		SET last_message_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	UpdateLastDM = `
		UPDATE direct_messages
		SET last_message_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
)

// Notification queries
const (
	CreateNotification = `
		INSERT INTO notifications (user_id, from_user_id, type, content,
		 related_id)
		VALUES (?, ?, ?, ?, ?)
	`

	GetUserNotifications = `
		SELECT id, from_user_id, type, content, is_read, related_id, created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	MarkNotificationsAsRead = `
		UPDATE notifications
		SET is_read = 1, read_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?
	`

	GetUnreadCount = `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = ? AND is_read = 0
	`
)

// Event queries
const (
	CreateEvent = `
		INSERT INTO events (uuid, group_id, creator_id,
		 title, description, event_time)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	CreateEventRSVP = `
		INSERT INTO event_rsvps (event_id, user_id, status)
		VALUES (?, ?, ?)
		ON CONFLICT(event_id, user_id) DO UPDATE SET status = ?
	`
)
