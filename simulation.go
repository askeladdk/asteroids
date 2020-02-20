package main

import (
	"image/color"
	"math/rand"

	"github.com/askeladdk/pancake/graphics"
	"github.com/askeladdk/pancake/graphics2d"
	"github.com/askeladdk/pancake/mathx"
)

const (
	ASTEROID = 1 << iota
	BULLET
	DELETED
	EPHEMERAL
	SPACESHIP
)

const SHIPID = 0

type actionCode int

const (
	FORWARD actionCode = iota
	TURN
)

type action struct {
	entityId int
	code     actionCode
	value    float32
}

type entity struct {
	sprite   graphics2d.Sprite // sprite
	pos      mathx.Vec2        // position
	vel      mathx.Vec2        // velocity
	rot      float32           // rotation
	rotv     float32           // rotational velocity
	dampenv  float32           // velocity dampening
	dampenr  float32           // rotational dampening
	maxv     float32           // maximum velocity
	minrotv  float32           // minimum rotational velocity
	turn     float32           // turn rate
	thrust   float32           // thrust speed
	mask     uint32            // capability mask
	radius   float32           // collision radius for COLLIDES
	lifetime float32           // time until death in seconds, for EPHEMERAL
	pos0     mathx.Vec2
	rot0     float32
}

type Simulation struct {
	sprites  []graphics2d.Sprite
	bounds   mathx.Rectangle
	entities []entity
	actions  []action
}

type entities struct {
	s *Simulation
	i int
	j int
	a float32
}

func (es *entities) Len() int {
	return es.j - es.i
}

func (es *entities) ColorAt(i int) color.NRGBA {
	return color.NRGBA{0xff, 0xff, 0xff, 0xff}
}

func (es *entities) at(i int) entity {
	return es.s.entities[es.i+i]
}

func (es *entities) Texture() *graphics.Texture {
	return es.at(0).sprite.Texture
}

func (es entities) TextureRegionAt(i int) mathx.Aff3 {
	return es.at(i).sprite.Region
}

func (es entities) ModelViewAt(i int) mathx.Aff3 {
	e := es.at(i)
	pos := e.pos0.Lerp(e.pos, es.a)
	rot := mathx.Lerp(e.rot0, e.rot, es.a)
	return mathx.
		ScaleAff3(e.sprite.Size).
		Rotated(rot).
		Translated(pos)
}

func (es entities) PivotAt(i int) mathx.Vec2 {
	return mathx.Vec2{}
}

func (s *Simulation) action(a action) {
	s.actions = append(s.actions, a)
}

func (s *Simulation) collisionResponse(a, b *entity) {
	if a.mask&b.mask&ASTEROID != 0 {
		v := a.pos.Sub(b.pos).Unit()
		a.vel = v.Mul(a.maxv * .5)
		b.vel = v.Mul(b.maxv * .5).Neg()
		a.rotv += mathx.Tau / 64 * (1 + 2*rand.Float32())
		b.rotv += mathx.Tau / 64 * (1 + 2*rand.Float32())
	} else if (a.mask|b.mask)&(ASTEROID|BULLET) == (ASTEROID | BULLET) {
		a.mask |= DELETED
		b.mask |= DELETED
	}
}

func (s *Simulation) collisionDetection() {
	for i := 0; i < len(s.entities); i++ {
		a := s.at(i)
		for j := i + 1; j < len(s.entities); j++ {
			b := s.at(j)
			c0 := mathx.Circle{a.pos, a.radius}
			c1 := mathx.Circle{b.pos, b.radius}
			if c0.IntersectsCircle(c1) {
				s.collisionResponse(a, b)
			}
		}
	}
}

func (s *Simulation) ephemeral(deltaTime float32) {
	for i, _ := range s.entities {
		e := s.at(i)
		if e.mask&EPHEMERAL != 0 {
			e.lifetime -= deltaTime
			if e.lifetime <= 0 {
				e.mask |= DELETED
			}
		}
	}
}

func (s *Simulation) deletePass() {
	count := len(s.entities)

	for i := 0; i < count; {
		if s.at(i).mask&DELETED != 0 {
			count--
			s.entities[i] = s.entities[count]
			s.entities = s.entities[:count]
		} else {
			i++
		}
	}
}

func (s *Simulation) processActions(dt float32) {
	for _, a := range s.actions {
		e := s.at(a.entityId)
		switch a.code {
		case FORWARD:
			acc := mathx.FromHeading(e.rot).Mul(a.value * e.thrust * dt)
			vel := e.vel.Add(acc)
			if vel.Len() > e.maxv {
				vel = vel.Unit().Mul(e.maxv)
			}
			e.vel = vel
		case TURN:
			e.rotv = e.turn * a.value * dt
		}
	}
	s.actions = s.actions[:0]
}

func (s *Simulation) frame(deltaTime float32) {
	s.processActions(deltaTime)
	s.collisionDetection()
	s.ephemeral(deltaTime)
	s.deletePass()

	for i, e := range s.entities {
		e.rot0 = e.rot
		e.pos0 = e.pos

		e.pos = e.pos.Add(e.vel.Mul(deltaTime))

		b := s.bounds.Expand(e.sprite.Size.Mul(0.5))
		if !e.pos.IntersectsRectangle(b) {
			e.pos = e.pos.Wrap(b)
			e.pos0 = e.pos
		}

		e.vel = e.vel.Mul(e.dampenv)
		e.rotv = mathx.Clamp(e.rotv*e.dampenr, -e.minrotv, e.minrotv)
		e.rot = e.rot + e.rotv*e.turn
		s.entities[i] = e
	}
}

func (s *Simulation) batches(alpha float32) []graphics2d.Batch {
	var batches []graphics2d.Batch

	for i := 0; i < len(s.entities); {
		spr := s.entities[i].sprite
		j := i + 1
		for ; j < len(s.entities); j++ {
			if s.entities[j].sprite != spr {
				break
			}
		}
		batches = append(batches, &entities{s, i, j, alpha})
		i = j
	}

	return batches
}

func (s *Simulation) at(i int) *entity {
	return &s.entities[i]
}

func (s *Simulation) spawnAsteroid() {
	pos := mathx.Vec2{
		rand.Float32(),
		rand.Float32(),
	}.MulVec2(s.bounds.Max)

	s.entities = append(s.entities, entity{
		sprite:  s.sprites[1],
		pos:     pos,
		turn:    mathx.Tau / 64 * (2*rand.Float32() - 1),
		maxv:    100,
		rotv:    1,
		minrotv: rand.Float32(),
		dampenr: 1,
		dampenv: 1,
		vel:     mathx.FromHeading(mathx.Tau * rand.Float32()).Mul(100),
		mask:    ASTEROID,
		radius:  28,
		pos0:    pos,
	})
}

func (s *Simulation) spawnBullet(pos mathx.Vec2, rot float32) {
	s.entities = append(s.entities, entity{
		sprite:   s.sprites[2],
		pos:      pos,
		dampenv:  1.01,
		rot:      rot,
		vel:      mathx.FromHeading(rot).Mul(200),
		mask:     EPHEMERAL | BULLET,
		radius:   4,
		lifetime: 0.6,
		pos0:     pos,
		rot0:     rot,
	})
}
