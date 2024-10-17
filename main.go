package main

import (
	"log"
	"time"

	"github.com/Peterculazh/block_fighter/pkg/game"
	"github.com/Peterculazh/block_fighter/pkg/tmp_storage"
	"github.com/gofrs/uuid"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"github.com/redis/go-redis/v9"
)

func main() {
	game.StartGame()
	app := iris.New()

	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Allow only your frontend origin
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"*"}, // Allow all headers
		AllowCredentials: true,
	})
	app.UseRouter(crs)
	app.Post("/join_game", join)

	app.Get("/join_game", func(ctx iris.Context) {
		socket, err := game.Upgrader.Upgrade(ctx.ResponseWriter(), ctx.Request())
		if err != nil {
			log.Printf("Accept: " + err.Error())
			return
		}
		socket.ReadLoop()
	})

	log.Println("Starting server on port 8080...")
	app.Listen(":8080")

}

func join(ctx iris.Context) {
	var body struct {
		Nickname string `json:"nickname"`
	}
	err := ctx.ReadJSON(&body)
	if err != nil {
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().Title("Incorrect nickname"))
		return
	}

	rdb := tmp_storage.GetRedis()
	// TODO: Add also checking for nickname in room before checking for redis for user in process
	val, err := rdb.Get(ctx.Request().Context(), body.Nickname).Result()
	switch {
	case err == redis.Nil:
	case err != nil:
		ctx.StopWithProblem(iris.StatusInternalServerError, iris.NewProblem().Title("Something went wrong").DetailErr(err))
		return
	case val != "":
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().Title("Nickname already taken."))
		return
	}

	println("Received nickname: " + body.Nickname)
	connectionTokenId, err := uuid.NewV4()
	if err != nil {
		ctx.StopWithProblem(iris.StatusInternalServerError, iris.NewProblem().Title("Something went wrong"))
		return
	}

	err = rdb.Set(ctx.Request().Context(), body.Nickname, connectionTokenId.String(), 10*time.Second).Err()
	if err != nil {
		ctx.StopWithProblem(iris.StatusInternalServerError, iris.NewProblem().Title("Something went wrong"))
		return
	}

	ctx.JSON(iris.Map{
		"status":   "success",
		"nickname": body.Nickname,
		"token":    connectionTokenId,
	})
}
