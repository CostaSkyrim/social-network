package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"social-network/backend/config"
	database "social-network/backend/db/sql"
	"social-network/backend/global"
)

// StartEventScheduler launches the background event-reminder worker. It runs
// until ctx is cancelled (the global shutdown context). The worker lives in the
// handlers package so it can reuse sendNotification and the WebSocket hub.
func StartEventScheduler(ctx context.Context, db *database.DataBase) {
	cfg := config.GetConfig()
	if cfg == nil || !cfg.Scheduler.Enabled {
		log.Println("Event scheduler disabled")
		return
	}

	tick := global.ParseDuration(cfg.Scheduler.TickInterval, time.Minute)
	lead := global.ParseDuration(cfg.Scheduler.ReminderLead, 30*time.Minute)
	if tick <= 0 {
		tick = time.Minute
	}
	if lead <= 0 {
		lead = 30 * time.Minute
	}

	log.Printf("Event scheduler started (tick=%s, reminder_lead=%s)", tick, lead)
	go eventSchedulerLoop(ctx, db, tick, lead)
}

func eventSchedulerLoop(ctx context.Context, db *database.DataBase, tick, lead time.Duration) {
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	// Run once at startup so reminders aren't delayed by a full tick.
	runEventReminders(ctx, db, lead)

	for {
		select {
		case <-ctx.Done():
			log.Println("Event scheduler stopped")
			return
		case <-ticker.C:
			runEventReminders(ctx, db, lead)
		}
	}
}

// runEventReminders sends a reminder for every event that starts within the lead
// window and hasn't been reminded yet, then marks it as handled.
func runEventReminders(ctx context.Context, db *database.DataBase, lead time.Duration) {
	now := time.Now().UTC()
	windowEnd := now.Add(lead)

	events, err := db.GetEventsDueForReminder(ctx, now, windowEnd)
	if err != nil {
		log.Printf("Event scheduler: failed to load due events: %v", err)
		return
	}

	for _, ev := range events {
		if ctx.Err() != nil {
			return
		}

		goingIDs, err := db.GetEventGoingUserIDs(ctx, ev.ID, ev.GroupID)
		if err != nil {
			log.Printf("Event scheduler: failed to load attendees for event %d: %v", ev.ID, err)
			continue
		}

		content := fmt.Sprintf("Reminder: '%s' in %s starts soon", ev.Title, ev.GroupTitle)
		for _, uid := range goingIDs {
			if uid == ev.CreatorID {
				continue
			}
			sendNotification(db, uid, ev.CreatorID, NotifEventReminder, content, &ev.ID, &ev.GroupUUID)
		}

		if err := db.MarkEventReminderSent(ctx, ev.ID); err != nil {
			log.Printf("Event scheduler: failed to mark event %d as reminded: %v", ev.ID, err)
			continue
		}

		log.Printf("Event scheduler: sent reminder for '%s' to %d attendees", ev.Title, len(goingIDs))
	}
}
