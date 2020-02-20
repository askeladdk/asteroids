package main

import (
	"image/color"
	"math/rand"

	"github.com/askeladdk/pancake/graphics"
	"github.com/askeladdk/pancake/mathx"
)

const (
	ASTEROID = 1 << iota
	BULLET
	DELETED
	EPHEMERAL
	SPACESHIP
)

const (
	ImageShip = iota
	ImageAsteroid
	ImageBullet
)

const SHIPID = 0

type ActionCode int

const (
	FORWARD ActionCode = iota
	TURN
)

type Action struct {
	EntityId int
	Code     ActionCode
	Value    float32
}

type Entity struct {
	ImageId  int        // image id
	Pos      mathx.Vec2 // position
	Vel      mathx.Vec2 // velocity
	Rot      float32    // rotation
	RotV     float32    // rotational velocity
	Acc      float32    // acceleration
	RotA     float32    // rotational acceleration
	MaxV     float32    // maximum velocity
	MinRotV  float32    // minimum rotational velocity
	Turn     float32    // turn rate
	Thrust   float32    // thrust speed
	Mask     uint32     // capability mask
	Radius   float32    // collision radius for COLLIDES
	Lifetime float32    // time until death in seconds, for EPHEMERAL
	Pos0     mathx.Vec2 // last position, for interpolation
	Rot0     float32    // last rotation, for interpolation
}

type Simulation struct {
	ImageAtlas *graphics.Texture
	Images     []graphics.Image
	Bounds     mathx.Rectangle
	Entities   []Entity
	Actions    []Action
	Alpha      float32
}

func (s *Simulation) Reset() {
	s.Entities = s.Entities[:0]
}

func (s *Simulation) Len() int {
	return len(s.Entities)
}

func (s *Simulation) ColorAt(i int) color.NRGBA {
	return color.NRGBA{0xff, 0xff, 0xff, 0xff}
}

func (s *Simulation) Texture() *graphics.Texture {
	return s.ImageAtlas
}

func (s *Simulation) TextureRegionAt(i int) graphics.TextureRegion {
	return s.Images[s.Entities[i].ImageId].TextureRegion()
}

func (s *Simulation) ModelViewAt(i int) mathx.Aff3 {
	e := s.At(i)
	pos := e.Pos0.Lerp(e.Pos, s.Alpha)
	rot := mathx.Lerp(e.Rot0, e.Rot, s.Alpha)
	return mathx.
		ScaleAff3(s.Images[e.ImageId].Scale()).
		Rotated(rot).
		Translated(pos)
}

func (s *Simulation) PivotAt(i int) mathx.Vec2 {
	return mathx.Vec2{}
}

func (s *Simulation) Action(entityId int, code ActionCode, value float32) {
	s.Actions = append(s.Actions, Action{entityId, code, value})
}

func (s *Simulation) collisionResponse(a, b *Entity) {
	if a.Mask&b.Mask&ASTEROID != 0 {
		v := a.Pos.Sub(b.Pos).Unit()
		a.Vel = v.Mul(a.MaxV * .5)
		b.Vel = v.Mul(b.MaxV * .5).Neg()
		a.RotV += mathx.Tau / 64 * (1 + 2*rand.Float32())
		b.RotV += mathx.Tau / 64 * (1 + 2*rand.Float32())
	} else if (a.Mask|b.Mask)&(ASTEROID|BULLET) == (ASTEROID | BULLET) {
		a.Mask |= DELETED
		b.Mask |= DELETED
	}
}

func (s *Simulation) processCollisions() {
	for i := 0; i < len(s.Entities); i++ {
		a := s.At(i)
		for j := i + 1; j < len(s.Entities); j++ {
			b := s.At(j)
			c0 := mathx.Circle{a.Pos, a.Radius}
			c1 := mathx.Circle{b.Pos, b.Radius}
			if c0.IntersectsCircle(c1) {
				s.collisionResponse(a, b)
			}
		}
	}
}

func (s *Simulation) processEphemeral(deltaTime float32) {
	for i, _ := range s.Entities {
		e := s.At(i)
		if e.Mask&EPHEMERAL != 0 {
			e.Lifetime -= deltaTime
			if e.Lifetime <= 0 {
				e.Mask |= DELETED
			}
		}
	}
}

func (s *Simulation) processDeletions() {
	count := len(s.Entities)

	for i := 0; i < count; {
		if s.At(i).Mask&DELETED != 0 {
			count--
			s.Entities[i] = s.Entities[count]
			s.Entities = s.Entities[:count]
		} else {
			i++
		}
	}
}

func (s *Simulation) processActions(dt float32) {
	for _, a := range s.Actions {
		e := s.At(a.EntityId)
		switch a.Code {
		case FORWARD:
			acc := mathx.FromHeading(e.Rot).Mul(a.Value * e.Thrust * dt)
			vel := e.Vel.Add(acc)
			if vel.Len() > e.MaxV {
				vel = vel.Unit().Mul(e.MaxV)
			}
			e.Vel = vel
		case TURN:
			e.RotV = e.Turn * a.Value * dt
		}
	}
	s.Actions = s.Actions[:0]
}

func (s *Simulation) Frame(deltaTime float32) {
	s.processActions(deltaTime)
	s.processCollisions()
	s.processEphemeral(deltaTime)
	s.processDeletions()

	for i, e := range s.Entities {
		e.Rot0 = e.Rot
		e.Pos0 = e.Pos

		e.Pos = e.Pos.Add(e.Vel.Mul(deltaTime))

		b := s.Bounds.Expand(s.Images[e.ImageId].Scale().Mul(0.5))
		if !e.Pos.IntersectsRectangle(b) {
			e.Pos = e.Pos.Wrap(b)
			e.Pos0 = e.Pos
		}

		e.Vel = e.Vel.Mul(e.Acc)
		e.RotV = mathx.Clamp(e.RotV*e.RotA, -e.MinRotV, e.MinRotV)
		e.Rot = e.Rot + e.RotV*e.Turn
		s.Entities[i] = e
	}
}

func (s *Simulation) At(i int) *Entity {
	return &s.Entities[i]
}

func (s *Simulation) SpawnAsteroid() {
	pos := mathx.Vec2{
		rand.Float32(),
		rand.Float32(),
	}.MulVec2(s.Bounds.Max)

	s.Entities = append(s.Entities, Entity{
		ImageId: ImageAsteroid,
		Pos:     pos,
		Turn:    mathx.Tau / 64 * (2*rand.Float32() - 1),
		MaxV:    100,
		RotV:    1,
		MinRotV: rand.Float32(),
		RotA:    1,
		Acc:     1,
		Vel:     mathx.FromHeading(mathx.Tau * rand.Float32()).Mul(100),
		Mask:    ASTEROID,
		Radius:  28,
		Pos0:    pos,
	})
}

func (s *Simulation) SpawnBullet(pos mathx.Vec2, rot float32) {
	s.Entities = append(s.Entities, Entity{
		ImageId:  ImageBullet,
		Pos:      pos,
		Acc:      1.01,
		Rot:      rot,
		Vel:      mathx.FromHeading(rot).Mul(200),
		Mask:     EPHEMERAL | BULLET,
		Radius:   4,
		Lifetime: 0.6,
		Pos0:     pos,
		Rot0:     rot,
	})
}

func (s *Simulation) SpawnSpaceship() {
	midscreen := s.Bounds.Max.Mul(0.5)
	s.Entities = append(s.Entities, Entity{
		ImageId: ImageShip,
		Pos0:    midscreen,
		Pos:     midscreen,
		Rot:     -mathx.Tau / 4,
		MinRotV: 1,
		MaxV:    300,
		Turn:    mathx.Tau / 4,
		Thrust:  100,
		RotA:    0.95,
		Acc:     0.99,
		Mask:    SPACESHIP,
		Radius:  14,
	})
}
