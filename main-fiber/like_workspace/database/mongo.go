package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Global ตัวแปรที่ใช้ทั้งโปรเจกต์
var Client *mongo.Client
var DB *mongo.Database

// ConnectMongo เชื่อมต่อ MongoDB และเซ็ตค่า Client, DB
func ConnectMongo() {
	// โหลดไฟล์ .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, fallback to environment vars")
	}

	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")

	if uri == "" || dbName == "" {
		log.Fatal("❌ MONGO_URI or DB_NAME not set in environment")
	}

	// สร้าง context พร้อม timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// เชื่อม MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("❌ Failed to connect MongoDB:", err)
	}

	// ping เช็คว่า connect ได้จริง
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal("❌ Failed to ping MongoDB:", err)
	}

	fmt.Println("✅ Connected to MongoDB!")

	Client = client
	DB = client.Database(dbName)
}

// DisconnectMongo ปิดการเชื่อมต่อ
func DisconnectMongo() {
	if Client != nil {
		if err := Client.Disconnect(context.TODO()); err != nil {
			log.Fatal("❌ Failed to disconnect MongoDB:", err)
		}
		fmt.Println("👋 Disconnected MongoDB!")
	}
}

// Helper function สำหรับ collection ต่าง ๆ
func Posts() *mongo.Collection          { return DB.Collection("posts") }
func PostCategories() *mongo.Collection { return DB.Collection("post_categories") }