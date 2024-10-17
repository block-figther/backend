package game

import (
	"github.com/ByteArena/box2d"
)

// EntityType enum
type EntityType int32

const (
	EntityGround EntityType = iota
	EntityPlayer
	EntityBall
)

// BaseEntityData interface
type BaseEntityData interface {
	GetType() EntityType
}

// Concrete types implementing BaseEntityData
type GroundData struct {
	Type   EntityType `json:"type"`
	Width  float64    `json:"width"`
	Height float64    `json:"height"`
}

type BallData struct {
	Type   EntityType `json:"type"`
	Radius float64    `json:"radius"`
}

type PlayerData struct {
	Type   EntityType `json:"type"`
	Width  float64    `json:"width"`
	Height float64    `json:"height"`
	Health int        `json:"health"`
	Id     string     `json:"-"`
}

// Implement GetType for each struct
func (g GroundData) GetType() EntityType { return g.Type }
func (b BallData) GetType() EntityType   { return b.Type }
func (p PlayerData) GetType() EntityType { return p.Type }

type WorldContainer struct {
	world box2d.B2World
}

var world = WorldContainer{}

func (wC *WorldContainer) CreateWorld() {
	wC.world = box2d.MakeB2World(box2d.MakeB2Vec2(0, -100))
}

func (wC *WorldContainer) CreatePlayer(x float64, y float64) *box2d.B2Body {
	bodyDef := box2d.MakeB2BodyDef()
	bodyDef.Type = box2d.B2BodyType.B2_dynamicBody
	bodyDef.Position.Set(x, y)
	width := 1.0
	height := 1.0

	box := wC.world.CreateBody(&bodyDef)
	// box.SetMassData(&box2d.B2MassData{
	// 	Mass: 10,
	// })

	shape := box2d.MakeB2PolygonShape()
	shape.SetAsBox(width, height)

	box.CreateFixture(&shape, 0)

	box.SetUserData(PlayerData{
		Width:  width,
		Height: height,
		Type:   EntityPlayer,
	})

	return box
}

func (wC *WorldContainer) CreateGround(w float64, h float64, x float64, y float64) *box2d.B2Body {
	bodyDef := box2d.MakeB2BodyDef()
	bodyDef.Position.Set(x, y)
	bodyDef.Type = box2d.B2BodyType.B2_staticBody

	wall := wC.world.CreateBody(&bodyDef)

	shape := box2d.MakeB2PolygonShape()
	shape.SetAsBox(w, h)

	wall.CreateFixture(&shape, 0)

	wall.SetUserData(GroundData{
		Width:  w,
		Height: h,
		Type:   EntityGround,
	})

	return wall
}

func (wC *WorldContainer) MovingPlayer(id string, direction PlayerMovingDirection) {
	for body := wC.world.GetBodyList(); body != nil; body.GetNext() {
		userData, ok := body.GetUserData().(PlayerData)
		if ok {
			if userData.Id == id {
				currentVelociy := body.GetLinearVelocity()
				switch direction {
				case up:
					body.SetLinearVelocity(box2d.B2Vec2{
						X: currentVelociy.X,
						Y: currentVelociy.Y + 50,
					})
				case left:
					body.SetLinearVelocity(box2d.B2Vec2{
						X: currentVelociy.X - 10,
						Y: currentVelociy.Y,
					})
				case down:
					body.SetLinearVelocity(box2d.B2Vec2{
						X: currentVelociy.X,
						Y: currentVelociy.Y - 10,
					})
				case right:
					body.SetLinearVelocity(box2d.B2Vec2{
						X: currentVelociy.X + 10,
						Y: currentVelociy.Y,
					})
				}
				break
			}
		}
	}
}
