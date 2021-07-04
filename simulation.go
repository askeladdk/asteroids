package main

import (
	"image/color"
	"math/rand"

	"github.com/askeladdk/pancake/graphics"
	"github.com/askeladdk/pancake/mathx"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

type gameState int

const (
	statePLAYING gameState = iota
	stateNEXTLEVEL
	stateGAMEOVER
)

const (
	flagASTEROID = 1 << iota
	flagBULLET
	flagDELETED
	flagEPHEMERAL
	flagSPACESHIP
	flagDEBRIS
)

const (
	imageShip = iota
	imageAsteroid
	imageBullet
	imageDebris0
	imageDebris1
	imageDebris2
	imageDebris3
)

const shipID = 0

type actionCode int

const (
	actionForward actionCode = iota
	actionTurn
	actionFire
)

type action struct {
	EntityID int
	Code     actionCode
	Value    float64
}

type entity struct {
	ImageID  int        // image id
	Pos      mathx.Vec2 // position
	Vel      mathx.Vec2 // velocity
	Rot      float64    // rotation
	RotV     float64    // rotational velocity
	Acc      float64    // acceleration
	RotA     float64    // rotational acceleration
	MaxV     float64    // maximum velocity
	MinRotV  float64    // minimum rotational velocity
	Turn     float64    // turn rate
	Thrust   float64    // thrust speed
	Mask     uint32     // capability mask
	Radius   float64    // collision radius for COLLIDES
	Lifetime float64    // time until death in seconds, for EPHEMERAL
	Pos0     mathx.Vec2 // last position, for interpolation
	Rot0     float64    // last rotation, for interpolation
}

type theSimulation struct {
	ImageAtlas *graphics.Texture
	Images     []graphics.Image
	Sounds     []*beep.Buffer
	Bounds     mathx.Rectangle
	Entities   []entity
	Actions    []action
	Alpha      float64
	State      gameState
	Level      int
	Score      int
	Remaining  int
}

var asteroidsPerLevel = []int{
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

func (s *theSimulation) PlaySound(i int) {
	snd := s.Sounds[i]
	speaker.Play(snd.Streamer(0, snd.Len()))
}

func (s *theSimulation) Reset() {
	s.State = statePLAYING
	s.Remaining = 0
	s.Entities = s.Entities[:0]
	s.SpawnSpaceship()
	for i := 0; i < asteroidsPerLevel[s.Level%len(asteroidsPerLevel)]; i++ {
		s.SpawnAsteroid()
	}
}

func (s *theSimulation) Len() int {
	return len(s.Entities)
}

func (s *theSimulation) TintColorAt(i int) color.Color {
	return color.RGBA{0xff, 0xff, 0xff, 0xff}
}

func (s *theSimulation) TextureAt(_ int) *graphics.Texture {
	return s.ImageAtlas
}

func (s *theSimulation) TextureRegionAt(i int) graphics.TextureRegion {
	return s.Images[s.Entities[i].ImageID].TextureRegion()
}

func (s *theSimulation) ModelViewAt(i int) mathx.Aff3 {
	e := s.At(i)
	pos := e.Pos0.Lerp(e.Pos, s.Alpha)
	rot := mathx.Lerp(e.Rot0, e.Rot, s.Alpha)
	return mathx.
		ScaleAff3(s.Images[e.ImageID].Scale()).
		Rotated(rot).
		Translated(pos)
}

func (s *theSimulation) OriginAt(i int) mathx.Vec2 {
	return mathx.Vec2{}
}

func (s *theSimulation) ZOrderAt(i int) float64 {
	return 0
}

func (s *theSimulation) Action(entityID int, code actionCode, value float64) {
	s.Actions = append(s.Actions, action{entityID, code, value})
}

func (s *theSimulation) collisionResponse(a, b *entity) {
	if a.Mask&(flagASTEROID|flagDEBRIS) != 0 && b.Mask&(flagASTEROID|flagDEBRIS) != 0 {
		v := a.Pos.Sub(b.Pos).Unit()
		a.Vel = v.Mul(a.MaxV * .5)
		b.Vel = v.Mul(b.MaxV * .5).Neg()
		a.RotV += mathx.Tau / 64 * (1 + 2*rand.Float64())
		b.RotV += mathx.Tau / 64 * (1 + 2*rand.Float64())
		s.PlaySound(2)
	} else if (a.Mask|b.Mask)&(flagASTEROID|flagBULLET) == (flagASTEROID | flagBULLET) {
		a.Mask |= flagDELETED
		b.Mask |= flagDELETED
		s.Score += 100
		s.Remaining--
		if a.Mask&flagASTEROID != 0 {
			s.SpawnDebris(a.Pos)
		} else {
			s.SpawnDebris(b.Pos)
		}
		s.PlaySound(1)
	} else if (a.Mask|b.Mask)&(flagDEBRIS|flagBULLET) == (flagDEBRIS | flagBULLET) {
		a.Mask |= flagDELETED
		b.Mask |= flagDELETED
		s.Score += 25
		s.Remaining--
		s.PlaySound(1)
	} else if a.Mask&flagSPACESHIP != 0 && b.Mask&(flagASTEROID|flagDEBRIS) != 0 {
		a.Mask |= flagDELETED
		s.State = stateGAMEOVER
		s.PlaySound(1)
	}
}

func (s *theSimulation) processCollisions() {
	for i := 0; i < len(s.Entities); i++ {
		a := s.At(i)
		for j := i + 1; j < len(s.Entities); j++ {
			b := s.At(j)
			c0 := mathx.Circle{Center: a.Pos, Radius: a.Radius}
			c1 := mathx.Circle{Center: b.Pos, Radius: b.Radius}
			if c0.IntersectsCircle(c1) {
				s.collisionResponse(a, b)
			}

			if a.Mask&flagDELETED != 0 {
				break
			}
		}
	}
}

func (s *theSimulation) processEphemeral(deltaTime float64) {
	for i := range s.Entities {
		e := s.At(i)
		if e.Mask&flagEPHEMERAL != 0 {
			e.Lifetime -= deltaTime
			if e.Lifetime <= 0 {
				e.Mask |= flagDELETED
			}
		}
	}
}

func (s *theSimulation) processDeletions() {
	count := len(s.Entities)

	for i := 0; i < count; {
		if s.At(i).Mask&flagDELETED != 0 {
			count--
			s.Entities[i] = s.Entities[count]
			s.Entities = s.Entities[:count]
		} else {
			i++
		}
	}
}

func (s *theSimulation) processActions(dt float64) {
	for _, a := range s.Actions {
		e := s.At(a.EntityID)
		switch a.Code {
		case actionForward:
			acc := mathx.FromHeading(e.Rot).Mul(a.Value * e.Thrust * dt)
			vel := e.Vel.Add(acc)
			if vel.Len() > e.MaxV {
				vel = vel.Unit().Mul(e.MaxV)
			}
			e.Vel = vel
		case actionTurn:
			e.RotV = e.Turn * a.Value * dt
		case actionFire:
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

func (s *theSimulation) processPhysics(deltaTime float64) {
	for i, e := range s.Entities {
		e.Rot0 = e.Rot
		e.Pos0 = e.Pos

		e.Pos = e.Pos.Add(e.Vel.Mul(deltaTime))

		b := s.Bounds.Expand(s.Images[e.ImageID].Scale().Mul(0.5))
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

func (s *theSimulation) Frame(deltaTime float64) {
	s.processActions(deltaTime)
	s.processCollisions()
	s.processEphemeral(deltaTime)
	s.processDeletions()
	s.processPhysics(deltaTime)

	if s.Remaining == 0 && s.State == statePLAYING {
		s.State = stateNEXTLEVEL
	}
}

func (s *theSimulation) At(i int) *entity {
	return &s.Entities[i]
}

func (s *theSimulation) SpawnAsteroid() {
	pos := s.Bounds.Max.
		Mul(.5).
		Add(mathx.FromHeading(mathx.Tau * rand.Float64()).Mul(128 + 128*rand.Float64()))

	s.Entities = append(s.Entities, entity{
		ImageID: imageAsteroid,
		Pos:     pos,
		Turn:    mathx.Tau / 64 * (2*rand.Float64() - 1),
		MaxV:    100,
		RotV:    1,
		MinRotV: rand.Float64(),
		RotA:    1,
		Acc:     1,
		Vel:     mathx.FromHeading(mathx.Tau * rand.Float64()).Mul(100),
		Mask:    flagASTEROID,
		Radius:  28,
		Pos0:    pos,
	})

	s.Remaining++
}

func (s *theSimulation) SpawnDebris(pos mathx.Vec2) {
	for i := 0; i < 4; i++ {
		heading := (mathx.Tau / 4) * float64(i)
		pos0 := pos.Add(mathx.FromHeading(heading).Mul(16))

		s.Entities = append(s.Entities, entity{
			ImageID: imageDebris0 + i,
			Pos:     pos0,
			Turn:    mathx.Tau / 32 * (2*rand.Float64() - 1),
			MaxV:    150,
			RotV:    1,
			MinRotV: rand.Float64(),
			RotA:    1,
			Acc:     1,
			Vel:     mathx.FromHeading(mathx.Tau * rand.Float64()).Mul(150),
			Mask:    flagDEBRIS,
			Radius:  14,
			Pos0:    pos0,
		})

		s.Remaining++
	}
}

func (s *theSimulation) SpawnBullet(pos mathx.Vec2, rot float64) {
	s.Entities = append(s.Entities, entity{
		ImageID:  imageBullet,
		Pos:      pos,
		Acc:      1.01,
		Rot:      rot,
		Vel:      mathx.FromHeading(rot).Mul(200),
		Mask:     flagEPHEMERAL | flagBULLET,
		Radius:   4,
		Lifetime: 0.6,
		Pos0:     pos,
		Rot0:     rot,
	})
}

func (s *theSimulation) SpawnSpaceship() {
	midscreen := s.Bounds.Max.Mul(0.5)
	s.Entities = append(s.Entities, entity{
		ImageID: imageShip,
		Pos0:    midscreen,
		Pos:     midscreen,
		Rot:     -mathx.Tau / 4,
		MinRotV: 1,
		MaxV:    300,
		Turn:    mathx.Tau / 4,
		Thrust:  100,
		RotA:    0.95,
		Acc:     0.99,
		Mask:    flagSPACESHIP,
		Radius:  14,
	})
}
