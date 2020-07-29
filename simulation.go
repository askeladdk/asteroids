package main

import (
	"image/color"
	"math/rand"

	"github.com/askeladdk/pancake/graphics"
	"github.com/askeladdk/pancake/mathx"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

type GameState int

const (
	PLAYING GameState = iota
	NEXTLEVEL
	GAMEOVER
)

const (
	ASTEROID = 1 << iota
	BULLET
	DELETED
	EPHEMERAL
	SPACESHIP
	DEBRIS
)

const (
	ImageShip = iota
	ImageAsteroid
	ImageBullet
	ImageDebris0
	ImageDebris1
	ImageDebris2
	ImageDebris3
)

const SHIPID = 0

type ActionCode int

const (
	FORWARD ActionCode = iota
	TURN
	FIRE
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
	Sounds     []*beep.Buffer
	Bounds     mathx.Rectangle
	Entities   []Entity
	Actions    []Action
	Alpha      float32
	State      GameState
	Level      int
	Score      int
	Remaining  int
}

var AsteroidsPerLevel = []int{
	1,
	2,
	3,
	5,
	8,
	13,
	21,
	34,
	55,
	89,
}

func (s *Simulation) PlaySound(i int) {
	snd := s.Sounds[i]
	speaker.Play(snd.Streamer(0, snd.Len()))
}

func (s *Simulation) Reset() {
	s.State = PLAYING
	s.Remaining = 0
	s.Entities = s.Entities[:0]
	s.SpawnSpaceship()
	for i := 0; i < AsteroidsPerLevel[s.Level%len(AsteroidsPerLevel)]; i++ {
		s.SpawnAsteroid()
	}
}

func (s *Simulation) Len() int {
	return len(s.Entities)
}

func (s *Simulation) TintColorAt(i int) color.NRGBA {
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

func (s *Simulation) OriginAt(i int) mathx.Vec2 {
	return mathx.Vec2{}
}

func (s *Simulation) Action(entityId int, code ActionCode, value float32) {
	s.Actions = append(s.Actions, Action{entityId, code, value})
}

func (s *Simulation) collisionResponse(a, b *Entity) {
	if a.Mask&(ASTEROID|DEBRIS) != 0 && b.Mask&(ASTEROID|DEBRIS) != 0 {
		v := a.Pos.Sub(b.Pos).Unit()
		a.Vel = v.Mul(a.MaxV * .5)
		b.Vel = v.Mul(b.MaxV * .5).Neg()
		a.RotV += mathx.Tau / 64 * (1 + 2*rand.Float32())
		b.RotV += mathx.Tau / 64 * (1 + 2*rand.Float32())
		s.PlaySound(2)
	} else if (a.Mask|b.Mask)&(ASTEROID|BULLET) == (ASTEROID | BULLET) {
		a.Mask |= DELETED
		b.Mask |= DELETED
		s.Score += 100
		s.Remaining--
		if a.Mask&ASTEROID != 0 {
			s.SpawnDebris(a.Pos)
		} else {
			s.SpawnDebris(b.Pos)
		}
		s.PlaySound(1)
	} else if (a.Mask|b.Mask)&(DEBRIS|BULLET) == (DEBRIS | BULLET) {
		a.Mask |= DELETED
		b.Mask |= DELETED
		s.Score += 25
		s.Remaining--
		s.PlaySound(1)
	} else if a.Mask&SPACESHIP != 0 && b.Mask&(ASTEROID|DEBRIS) != 0 {
		a.Mask |= DELETED
		s.State = GAMEOVER
		s.PlaySound(1)
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

			if a.Mask&DELETED != 0 {
				break
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
		case FIRE:
			s.SpawnBullet(e.Pos, e.Rot)
			s.PlaySound(0)
			s.Score -= 5
			if s.Score < 0 {
				s.Score = 0
			}
		}
	}
	s.Actions = s.Actions[:0]
}

func (s *Simulation) processPhysics(deltaTime float32) {
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

func (s *Simulation) Frame(deltaTime float32) {
	s.processActions(deltaTime)
	s.processCollisions()
	s.processEphemeral(deltaTime)
	s.processDeletions()
	s.processPhysics(deltaTime)

	if s.Remaining == 0 && s.State == PLAYING {
		s.State = NEXTLEVEL
	}
}

func (s *Simulation) At(i int) *Entity {
	return &s.Entities[i]
}

func (s *Simulation) SpawnAsteroid() {
	pos := s.Bounds.Max.
		Mul(.5).
		Add(mathx.FromHeading(mathx.Tau * rand.Float32()).Mul(128 + 128*rand.Float32()))

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

	s.Remaining++
}

func (s *Simulation) SpawnDebris(pos mathx.Vec2) {
	for i := 0; i < 4; i++ {
		heading := (mathx.Tau / 4) * float32(i)
		pos0 := pos.Add(mathx.FromHeading(heading).Mul(16))

		s.Entities = append(s.Entities, Entity{
			ImageId: ImageDebris0 + i,
			Pos:     pos0,
			Turn:    mathx.Tau / 32 * (2*rand.Float32() - 1),
			MaxV:    150,
			RotV:    1,
			MinRotV: rand.Float32(),
			RotA:    1,
			Acc:     1,
			Vel:     mathx.FromHeading(mathx.Tau * rand.Float32()).Mul(150),
			Mask:    DEBRIS,
			Radius:  14,
			Pos0:    pos0,
		})

		s.Remaining++
	}
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
