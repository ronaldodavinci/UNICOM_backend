package routes

import (
    "fmt"
    "github.com/gofiber/fiber/v2"
    "go.mongodb.org/mongo-driver/v2/mongo"
)

func GetUsersHandler(client *mongo.Client) fiber.Handler {
    return func(c *fiber.Ctx) error {
        collection := client.Database("test").Collection("Posts")

        cursor, err := collection.Find(c.Context(), map[string]interface{}{})
        if err != nil {
            return c.Status(500).SendString(err.Error())
        }
        defer cursor.Close(c.Context())

        var results []map[string]interface{}
        if err := cursor.All(c.Context(), &results); err != nil {
            return c.Status(500).SendString(err.Error())
        }

        fmt.Println("📦 Results from Mongo:", results) // debug log

        return c.JSON(results)
    }	
}
