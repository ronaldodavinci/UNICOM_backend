package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"main-webbase/config"
	"main-webbase/database"
	"main-webbase/internal/services"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.MongoURI == "" || cfg.MongoURI == "mongodf://localhost:27017" {
		log.Fatal("please set MONGO_URI to point to your MongoDB instance")
	}

	client := database.ConnectMongo(cfg.MongoURI, cfg.MongoDB)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = client.Disconnect(ctx)
	}()

	db := client.Database(cfg.MongoDB)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}

	eventID := bson.NewObjectID()
	userID := bson.NewObjectID()
	scheduleID := bson.NewObjectID()
	regID := bson.NewObjectID()

	now := time.Now()
	firstSlot := now.Add(2 * time.Hour).UTC()
	slotEnd := firstSlot.Add(90 * time.Minute)

	eventsCol := db.Collection("events")
	schedCol := db.Collection("event_schedules")
	regCol := db.Collection("event_registrations")
	notiCol := db.Collection("notifications")

	cleanup := func(ctx context.Context) {
		_, _ = notiCol.DeleteMany(ctx, bson.M{
			"type":         services.NotiEventReminder,
			"ref.event_id": eventID,
			"user_id":      userID,
		})
		_, _ = regCol.DeleteMany(ctx, bson.M{"_id": regID})
		_, _ = schedCol.DeleteMany(ctx, bson.M{"_id": scheduleID})
		_, _ = eventsCol.DeleteMany(ctx, bson.M{"_id": eventID})
	}
	defer cleanup(ctx)

	if _, err := eventsCol.InsertOne(ctx, bson.M{
		"_id":    eventID,
		"title":  "Reminder test event",
		"topic":  "Reminder test event",
		"status": "active",
	}); err != nil {
		log.Fatalf("failed to insert test event: %v", err)
	}

	if _, err := schedCol.InsertOne(ctx, bson.M{
		"_id":        scheduleID,
		"event_id":   eventID,
		"date":       firstSlot,
		"time_start": firstSlot,
		"time_end":   slotEnd,
	}); err != nil {
		log.Fatalf("failed to insert test schedule: %v", err)
	}

	if _, err := regCol.InsertOne(ctx, bson.M{
		"_id":      regID,
		"event_id": eventID,
		"user_id":  userID,
		"status":   "registered",
	}); err != nil {
		log.Fatalf("failed to insert test registration: %v", err)
	}

	if _, err := notiCol.DeleteMany(ctx, bson.M{
		"type":         services.NotiEventReminder,
		"ref.event_id": eventID,
		"user_id":      userID,
	}); err != nil {
		log.Fatalf("failed to clear existing notifications: %v", err)
	}

	if err := services.RunEventReminder(ctx, db, loc, 14); err != nil {
		log.Fatalf("RunEventReminder returned error: %v", err)
	}

	var result bson.M
	err = notiCol.FindOne(ctx, bson.M{
		"type":         services.NotiEventReminder,
		"ref.event_id": eventID,
		"user_id":      userID,
	}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("no reminder notification found for the test data")
			os.Exit(1)
		}
		log.Fatalf("failed to fetch reminder notification: %v", err)
	}

	fmt.Println("✅ Event reminder notification inserted:")
	for k, v := range result {
		fmt.Printf("  %s: %#v\n", k, v)
	}
}
