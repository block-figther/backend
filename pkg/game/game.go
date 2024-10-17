package game

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/OakmoundStudio/oak/timing"
)

const (
	velocityIterations = 360
	positionIterations = 180
)

// BodyData struct to hold common data
type BodyData struct {
	X    float64        `json:"x"`
	Y    float64        `json:"y"`
	Type EntityType     `json:"type"`
	Data BaseEntityData `json:"data"`
}

type Game struct {
	worldContainer *WorldContainer
}

var game = Game{}

func StartGame() {
	world.CreateWorld()
	game.worldContainer = &world

	world.CreateGround(30, 0.4, 0, 0)

	go game.runWorldLoop()
}

func (game *Game) sendPlayerFrame(id string) {
	bodies := make([]BodyData, 0)

	for body := game.worldContainer.world.GetBodyList(); body != nil; body = body.GetNext() {
		pos := body.GetPosition()
		user_data := body.GetUserData()

		if user_data == nil {
			continue
		}

		var bodyData BodyData
		bodyData.X = pos.X
		bodyData.Y = pos.Y

		switch data := user_data.(type) {
		case GroundData:
			bodyData.Type = data.Type
			bodyData.Data = data
		case BallData:
			bodyData.Type = data.Type
			bodyData.Data = data
		case PlayerData:
			bodyData.Type = data.Type
			bodyData.Data = data
		default:
			fmt.Printf("Unknown user data type: %T\n", data)
			continue
		}

		bodies = append(bodies, bodyData)
	}

	data, err := json.Marshal(bodies)
	if err != nil {
		return
	}
	communication.SendMessage(id, data)
}

func (game *Game) runWorldLoop() {
	const targetFPS = 60
	targetFrameDuration := time.Second / targetFPS

	frameCount := 0
	lastFPSCheck := time.Now()

	// Create a DynamicTicker
	fpsTicker := timing.NewDynamicTicker()
	fpsTicker.SetTick(targetFrameDuration)

	for {
		start := time.Now()

		// Step the physics simulation
		// For fixed time step:

		game.worldContainer.world.Step(targetFrameDuration.Seconds(), velocityIterations, positionIterations)

		ids := room.GetPlayersIds()
		for _, id := range ids {
			game.sendPlayerFrame(id)
		}

		// Increment frame count
		frameCount++

		// Calculate and log FPS every second
		if time.Since(lastFPSCheck) >= time.Second {
			fps := float64(frameCount) / time.Since(lastFPSCheck).Seconds()
			log.Printf("FPS: %.2f", fps)
			frameCount = 0
			lastFPSCheck = time.Now()
		}

		// Calculate frame duration
		frameDuration := time.Since(start)

		// Log if frame took too long
		if frameDuration > targetFrameDuration {
			log.Printf("Frame took longer than expected: %v", frameDuration)
		}

		// Wait for next tick
		<-fpsTicker.C
	}
}

func (game *Game) AddPlayer(id string) {
	box := world.CreatePlayer(0, 200)
	userData, ok := box.M_userData.(PlayerData)
	if !ok {
		panic(`Test`)
	}

	userData.Id = id
	userData.Health = 100
	box.SetUserData(userData)
}

func (game *Game) RemovePlayer(id string) {
	for body := game.worldContainer.world.GetBodyList(); body != nil; body = body.GetNext() {
		userData, ok := body.GetUserData().(PlayerData)
		if ok {
			log.Printf("Player id %s, target id %s", userData.Id, id)
			if userData.Id == id {
				world.world.DestroyBody(body)
				break
			}
		}
	}
}

/**
* TODO: Rework into other system that handles keyboard clicks and control what system to call on key event
* and get rid of world's dependency on player moving direction or other keyboard event
 */
type PlayerMovingDirection string

const (
	up    PlayerMovingDirection = "up"
	left  PlayerMovingDirection = "left"
	down  PlayerMovingDirection = "down"
	right PlayerMovingDirection = "right"
)

func (game *Game) HandleKeyPress(id string, direction PlayerMovingDirection) {
	game.worldContainer.MovingPlayer(id, direction)
}
