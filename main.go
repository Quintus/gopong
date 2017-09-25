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
#cgo LDFLAGS: -lallegro_ttf -lallegro_font -lallegro_image -lallegro_primitives -lallegro
#include <stdio.h>
#include <stdlib.h>
#include <allegro5/allegro.h>
#include <allegro5/allegro_primitives.h>
#include <allegro5/allegro_image.h>
#include <allegro5/allegro_font.h>
#include <allegro5/allegro_ttf.h>

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
import (
	"fmt"
	"unsafe"
)

const WINDOWWIDTH int = 640
const WINDOWHEIGHT int = 480
const SPEED float32 = 2.0
var game_font *C.ALLEGRO_FONT

type Player struct {
	X float32
	Y float32
	W float32
	H float32
	Speed float32
	Points int
	origX float32
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

func NewPlayer(x float32) *Player {
	var player Player
	player.W = 10.0
	player.H = 50.0
	player.origX = x
	player.Reset()
	return &player
}

func (self *Player) Reset() {
	self.X = self.origX
	self.Y = float32(WINDOWHEIGHT) * 0.5 - self.H * 0.5
	self.Points = 0
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
	var ball Ball
	ball.Radius = 10.0
	ball.Xspeed = SPEED
	ball.Yspeed = SPEED
	ball.Reset()
	return &ball
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
		C.al_map_rgb(255, 255, 0))
}

func (self *Ball) Reset() {
	self.X = float32(WINDOWWIDTH) * 0.5
	self.Y = float32(WINDOWHEIGHT) * 0.5
	self.Xspeed = SPEED
	self.Yspeed = SPEED
}

func (self *Ball) Turn() {
	self.Xspeed *= -1.1
	self.Yspeed *= 1.1
}

func main() {
	C.al_install_system(C.ALLEGRO_VERSION_INT, nil)
	C.al_install_keyboard()
	C.al_init_primitives_addon()
	C.al_init_image_addon()
	C.al_init_font_addon()
	C.al_init_ttf_addon()
	defer C.al_uninstall_system()

	fmt.Println("Start")

	p_display := C.al_create_display(C.int(WINDOWWIDTH), C.int(WINDOWHEIGHT))
	p_evqueue := C.al_create_event_queue()

	defer C.al_destroy_display(p_display)
	defer C.al_destroy_event_queue(p_evqueue)

	C.al_register_event_source(p_evqueue, C.al_get_display_event_source(p_display))
	C.al_register_event_source(p_evqueue, C.al_get_keyboard_event_source())

	fontname := C.CString("DejaVuSansMono.ttf")
	game_font = C.al_load_ttf_font(fontname, 20, 0)
	defer C.al_destroy_font(game_font)
	C.free(unsafe.Pointer(fontname))

	mainloop(p_evqueue)

	fmt.Println("Finish")
}

func mainloop(p_evqueue *C.ALLEGRO_EVENT_QUEUE) {
	player1 := NewPlayer(50.0)
	player2 := NewPlayer(float32(WINDOWWIDTH) - 50.0)
	ball := NewBall()

	game_objects := [...]GameObject{ player1, player2, ball }

	run := true
	for run {
		// Handle events
		var evt C.ALLEGRO_EVENT
		for C.al_get_next_event(p_evqueue, &evt) {
			switch (C.extract_event(&evt)) {
			case C.ALLEGRO_EVENT_DISPLAY_CLOSE:
				run = false
			case C.ALLEGRO_EVENT_KEY_DOWN:
				handle_key_down(C.extract_keyboard_event(&evt).keycode, player1, player2, ball, &run);
			case C.ALLEGRO_EVENT_KEY_UP:
				handle_key_up(C.extract_keyboard_event(&evt).keycode, player1, player2, &run);
			}
		}

		// Clear screen
		C.al_clear_to_color(C.al_map_rgb(0, 0, 0))

		update(player1, player2, ball, game_objects[:])
		draw(player1, player2, game_objects[:])

		// Switch buffers
		C.al_flip_display()
	}
}

func update(player1 *Player, player2 *Player, ball *Ball, game_objects []GameObject) {
	for _, obj := range game_objects {
		obj.Update()
	}

	check_collisions(player1, player2, ball)
}

func draw (player1 *Player, player2 *Player, game_objects []GameObject) {
	for _, obj := range game_objects {
		obj.Draw()
	}

	points1 := fmt.Sprintf("%04d", player1.Points)
	points2 := fmt.Sprintf("%04d", player2.Points)

	cpoints1 := C.CString(points1)
	cpoints2 := C.CString(points2)
	defer C.free(unsafe.Pointer(cpoints1))
	defer C.free(unsafe.Pointer(cpoints2))

	C.al_draw_text(game_font, C.al_map_rgb(255, 255, 255), 10, 10, 0, cpoints1)
	C.al_draw_text(game_font, C.al_map_rgb(255, 255, 255), C.float(WINDOWWIDTH - 10), 10, C.ALLEGRO_ALIGN_RIGHT, cpoints2)
}

func check_collisions(player1 *Player, player2 *Player, ball *Ball) {
	// Return ball and speed it up
	for x := ball.X - 0.5 * ball.Radius; x < ball.X + 0.5 * ball.Radius; x++ {
		if (x >= player1.X && x < player1.X + player1.W) && (ball.Y >= player1.Y && ball.Y < player1.Y + player1.H) {
			ball.Turn()
			break
		} else if (x >= player2.X && x < player2.X + player2.W) && (ball.Y >= player2.Y && ball.Y < player2.Y + player2.H) {
			ball.Turn()
			break
		}
	}

	if ball.X <= 0 {
		player2.Points += 1

		ball.Reset()
	} else if ball.X + ball.Radius * 0.5 >= float32(WINDOWWIDTH) {
		player1.Points += 1

		ball.Reset()
	}
}

func handle_key_down(keycode C.int, player1 *Player, player2 *Player, ball *Ball, run *bool) {
	switch (keycode) {
	case C.ALLEGRO_KEY_ESCAPE:
		*run = false
	case C.ALLEGRO_KEY_UP:
		player2.Speed = -SPEED
	case C.ALLEGRO_KEY_DOWN:
		player2.Speed = SPEED
	case C.ALLEGRO_KEY_W:
		player1.Speed = -SPEED
	case C.ALLEGRO_KEY_S:
		player1.Speed = SPEED
	case C.ALLEGRO_KEY_ENTER:
		player1.Reset()
		player2.Reset()
		ball.Reset()
	}
}

func handle_key_up(keycode C.int,  player1 *Player, player2 *Player, run *bool) {
	switch keycode {
	case C.ALLEGRO_KEY_UP, C.ALLEGRO_KEY_DOWN:
		player2.Speed = 0
	case C.ALLEGRO_KEY_W, C.ALLEGRO_KEY_S:
		player1.Speed = 0
	}
}
