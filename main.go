/*
A pong game in Go for practising with Go.
Copyright (C) 2017  Marvin Gülker

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
package main

/*
#cgo LDFLAGS: -lallegro -lallegro_primitives
#include <stdio.h>
#include <stdlib.h>
#include <allegro5/allegro.h>
#include <allegro5/allegro_primitives.h>

inline ALLEGRO_EVENT_TYPE extract_event(const ALLEGRO_EVENT* p_evt)
{
  return p_evt->type;
}

inline const ALLEGRO_KEYBOARD_EVENT* extract_keyboard_event(const ALLEGRO_EVENT* p_evt)
{
  return &(p_evt->keyboard);
}
*/
import "C"
import "fmt"

const WINDOWWIDTH int = 640
const WINDOWHEIGHT int = 480
const SPEED float32 = 2.0

type Player struct {
	X float32
	Y float32
	W float32
	H float32
	Speed float32
	Points int
}

type Ball struct {
	X float32
	Y float32
	Radius float32
	Xspeed float32
	Yspeed float32
}

type GameObject interface {
	Update()
	Draw()
}

func NewPlayer() *Player {
	return &Player{0.0, 0.0, 5.0, 20.0, 0.0, 0}
}

func (self *Player) Update() {
	self.Move(self.Speed)
}

func (self *Player) Draw() {
	C.al_draw_filled_rectangle(
		C.float(self.X),
		C.float(self.Y),
		C.float(self.X + self.W),
		C.float(self.Y + self.H),
		C.al_map_rgb(255, 0, 0))
}

func (self *Player) Move(delta float32) {
	self.Y += delta

	if self.Y < 0 {
		self.Y = 0
	} else if self.Y + self.H > float32(WINDOWHEIGHT) {
		self.Y = float32(WINDOWHEIGHT) - self.H
	}
}

func NewBall() *Ball {
	return &Ball{0.0, 0.0, 5.0, SPEED, SPEED}
}

func (self *Ball) Update() {
	self.X += self.Xspeed
	self.Y += self.Yspeed

	switch {
	case self.Y - self.Radius <= 0:
		self.Yspeed = -self.Yspeed
	case self.Y + self.Radius >= float32(WINDOWHEIGHT):
		self.Yspeed = -self.Yspeed
	}
}

func (self *Ball) Draw() {
	C.al_draw_filled_circle(
		C.float(self.X),
		C.float(self.Y),
		C.float(self.Radius),
		C.al_map_rgb(0, 0, 255))
}

func main() {
	C.al_install_system(C.ALLEGRO_VERSION_INT, nil)
	C.al_install_keyboard()
	C.al_init_primitives_addon()
	defer C.al_uninstall_system()

	fmt.Println("Start")

	p_display := C.al_create_display(C.int(WINDOWWIDTH), C.int(WINDOWHEIGHT))
	p_evqueue := C.al_create_event_queue()

	defer C.al_destroy_display(p_display)
	defer C.al_destroy_event_queue(p_evqueue)

	C.al_register_event_source(p_evqueue, C.al_get_display_event_source(p_display))
	C.al_register_event_source(p_evqueue, C.al_get_keyboard_event_source())

	mainloop(p_evqueue)

	fmt.Println("Finish")
}

func mainloop(p_evqueue *C.ALLEGRO_EVENT_QUEUE) {
	player1 := NewPlayer()
	player2 := NewPlayer()
	ball := NewBall()

	game_objects := [...]GameObject{ player1, player2, ball }

	player1.X = 100
	player1.Y = 100

	player2.X = 400
	player2.Y = 100

	ball.X = 200
	ball.Y = 200

	run := true
	for run {
		// Handle events
		var evt C.ALLEGRO_EVENT
		for C.al_get_next_event(p_evqueue, &evt) {
			switch (C.extract_event(&evt)) {
			case C.ALLEGRO_EVENT_DISPLAY_CLOSE:
				run = false
			case C.ALLEGRO_EVENT_KEY_DOWN:
				handle_key_down(C.extract_keyboard_event(&evt).keycode, player1, player2, &run);
			case C.ALLEGRO_EVENT_KEY_UP:
				handle_key_up(C.extract_keyboard_event(&evt).keycode, player1, player2, &run);
			}
		}

		// Clear screen
		C.al_clear_to_color(C.al_map_rgb(0, 0, 0))

		update(game_objects[:])
		check_collisions(player1, player2, ball)
		draw(game_objects[:])

		// Switch buffers
		C.al_flip_display()
	}
}

func update(game_objects []GameObject) {
	for _, obj := range game_objects {
		obj.Update()
	}
}

func draw (game_objects []GameObject) {
	for _, obj := range game_objects {
		obj.Draw()
	}
}

func check_collisions(player1 *Player, player2 *Player, ball *Ball) {
	for x := ball.X - 0.5 * ball.Radius; x < ball.X + 0.5 * ball.Radius; x++ {
		if (x >= player1.X && x < player1.X + player1.W) && (ball.Y >= player1.Y && ball.Y < player1.Y + player1.H) {
			ball.Xspeed = -ball.Xspeed
			break
		} else if (x >= player2.X && x < player2.X + player2.W) && (ball.Y >= player2.Y && ball.Y < player2.Y + player2.H) {
			ball.Xspeed = -ball.Xspeed
			break
		}
	}

	if ball.X <= 0 {
		player2.Points += 1
		ball.X = float32(WINDOWWIDTH) * 0.5
		ball.Y = float32(WINDOWHEIGHT) * 0.5
	} else if ball.X + ball.Radius * 0.5 >= float32(WINDOWWIDTH) {
		ball.X = float32(WINDOWWIDTH) * 0.5
		ball.Y = float32(WINDOWHEIGHT) * 0.5
		player1.Points += 1
	}
}

func handle_key_down(keycode C.int, player1 *Player, player2 *Player, run *bool) {
	switch (keycode) {
	case C.ALLEGRO_KEY_ESCAPE:
		*run = false
	case C.ALLEGRO_KEY_UP:
		player1.Speed = -SPEED
	case C.ALLEGRO_KEY_DOWN:
		player1.Speed = SPEED
	case C.ALLEGRO_KEY_W:
		player2.Speed = -SPEED
	case C.ALLEGRO_KEY_S:
		player2.Speed = SPEED
	}
}

func handle_key_up(keycode C.int,  player1 *Player, player2 *Player, run *bool) {
	switch keycode {
	case C.ALLEGRO_KEY_UP, C.ALLEGRO_KEY_DOWN:
		player1.Speed = 0
	case C.ALLEGRO_KEY_W, C.ALLEGRO_KEY_S:
		player2.Speed = 0
	}
}
