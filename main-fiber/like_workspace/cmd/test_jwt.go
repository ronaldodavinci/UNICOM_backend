package main

import (
    "fmt"
    "log"
    "os"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/joho/godotenv"
)

func init() {
    // โหลดค่าจาก .env ถ้ามี
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using system environment variables")
    }
    log.Printf("env: JWT_SECRET len=%d", len(os.Getenv("JWT_SECRET")))

}

func main() {
    // ใช้ secret key จาก ENV หรือ fallback
    log.Printf("env: JWT_SECRET len=%d", len(os.Getenv("JWT_SECRET")))

    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        panic("JWT_SECRET is required")
    }

    // กำหนด claims
    claims := jwt.MapClaims{
        "sub": "68bf0f1a2a3c4d5e6f708091",       // user id
        "exp": time.Now().Add(time.Hour).Unix(), // หมดอายุใน 1 ชม.
    }

    // สร้าง token ด้วย HS256
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // เซ็นด้วย secret
    signed, err := token.SignedString([]byte(secret))
    if err != nil {
        panic(err)
    }

    fmt.Println("JWT =", signed)
}
