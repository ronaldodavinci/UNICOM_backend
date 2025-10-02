package routes

import (
	"main-webbase/internal/controllers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupAuth(app *fiber.App, client *mongo.Client) {
	app.Post("/register", func(c *fiber.Ctx) error {
		return controllers.Register(c, client)
	})
	// ตัวอย่าง
	//                   curl
	// curl -X POST http://127.0.0.1:8000/register \
	// -H "Content-Type: application/json" \
	// -d '{
	//     "firstname": "Jaiyok",
	//     "lastname": "Jaiyuk",
	//     "thaiprefix": "นางสาว",
	//     "gender": "Female",
	//     "type_person": "student",
	//     "student_id": "6610505678",
	//     "advisor_id": "123",
	//     "email": "kirana.jy@ku.ac.th",
	//     "password": "Jaiyok"
	// }'

	//.                 POSTMAN
	// URL: http://127.0.0.1:8000/register
	// Method: POST
	// Headers:
	// Key: Content-Type, Value: application/json
	// Body: raw, JSON
	// {
	//   "firstName": "Alice",
	//   "lastName": "Smith",
	//   "thaiprefix": "นางสาว",
	//   "gender": "Female",
	//   "type_person": "student",
	//   "student_id": "65012345",
	//   "advisor_id": "123",
	//   "email": "alice@example.com",
	//   "password": "yourpassword"
	// }
	app.Post("/login", func(c *fiber.Ctx) error {
		return controllers.Login(c, client)
	})
	// ตัวอย่าง
	// curl -X POST http://127.0.0.1:8000/login \
	// -H "Content-Type: application/json" \
	// -d '{
	// "email": "alice@example.com",
	// "password": "yourpassword"
	// }'
}









// list of data 7/9/2025 in big_workspace -> user


// curl -X POST http://127.0.0.1:8000/register \
//     -H "Content-Type: application/json" \
//     -d '{
//         "firstname": "Jaiyok",
//         "lastname": "Jaiyuk",
//         "thaiprefix": "นางสาว",
//         "gender": "Female",
//         "type_person": "student",
//         "student_id": "6610505678",
//         "advisor_id": "123",
//         "email": "kirana.jy@ku.ac.th",
//         "password": "Jaiyok"
//     }'

// curl -X POST http://127.0.0.1:8000/register \
//     -H "Content-Type: application/json" \
//     -d '{
//         "firstname": "Jittat",
//         "lastname": "Fakcharoen",
//         "thaiprefix": "นาย",
//         "gender": "Male",
//         "type_person": "advisor",
//         "student_id": "",
//         "advisor_id": "123",
//         "email": "jittat.jt@ku.ac.th",
//         "password": "Jittat"
//     }'

// curl -X POST http://127.0.0.1:8000/register \
//     -H "Content-Type: application/json" \
//     -d '{
//         "firstname": "Patompong",
//         "lastname": "Baworncharoenpun",
//         "thaiprefix": "นาย",
//         "gender": "Male",
//         "type_person": "student",
//         "student_id": "6610505462",
//         "advisor_id": "123",
//         "email": "Patompong.b@ku.ac.th",
//         "password": "password"
//     }'
