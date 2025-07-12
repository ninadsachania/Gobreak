package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

const FPS = 60
const DELTA_TIME_SEC = 1.0 / FPS
const WINDOW_WIDTH = 800
const WINDOW_HEIGHT = 600
const BACKGROUND_COLOR = 0xFF181818
const PROJ_SIZE = 25 * 0.80
const PROJ_SPEED = 350
const PROJ_COLOR = 0xFFFFFFFF
const BAR_LEN = 100
const BAR_THICCNESS = PROJ_SIZE
const BAR_Y = WINDOW_HEIGHT - PROJ_SIZE - 50
const BAR_SPEED = PROJ_SPEED * 1.5
const BAR_COLOR = 0xFF3030FF
const TARGET_WIDTH = BAR_LEN
const TARGET_HEIGHT = PROJ_SIZE
const TARGET_PADDING_X = 20
const TARGET_PADDING_Y = 50
const TARGET_ROWS = 4
const TARGET_COLS = 5
const TARGET_GRID_WIDTH = (TARGET_COLS*TARGET_WIDTH + (TARGET_COLS-1)*TARGET_PADDING_X)
const TARGET_GRID_X = WINDOW_WIDTH/2 - TARGET_GRID_WIDTH/2
const TARGET_GRID_Y = 50
const TARGET_COLOR = 0xFF30FF30

type target struct {
	x    float32
	y    float32
	dead bool
}

func init_targets() [TARGET_ROWS * TARGET_COLS]target {
	var targets [TARGET_ROWS * TARGET_COLS]target
	for row := 0; row < TARGET_ROWS; row += 1 {
		for col := 0; col < TARGET_COLS; col += 1 {
			targets[row*TARGET_COLS+col] = target{x: TARGET_GRID_X + (TARGET_WIDTH+TARGET_PADDING_X)*float32(col), y: TARGET_GRID_Y + TARGET_PADDING_Y*float32(row)}
		}
	}
	return targets
}

var targets_pool = init_targets()
var bar_x float32 = WINDOW_WIDTH/2 - BAR_LEN/2
var bar_dx float32 = 0
var proj_x float32 = WINDOW_WIDTH/2 - PROJ_SIZE/2
var proj_y float32 = BAR_Y - BAR_THICCNESS/2 - PROJ_SIZE
var proj_dx float32 = 1
var proj_dy float32 = 1
var quit = false
var pause = false
var started = false

// TODO: death
// TODO: score
// TODO: victory

func make_rect(x float32, y float32, w float32, h float32) sdl.Rect {
	return sdl.Rect{X: int32(x), Y: int32(y), W: int32(w), H: int32(h)}
}

func set_color(renderer *sdl.Renderer, color uint32) {
	var r = uint8((color >> (0 * 8)) & 0xFF)
	var g = uint8((color >> (1 * 8)) & 0xFF)
	var b = uint8((color >> (2 * 8)) & 0xFF)
	var a = uint8((color >> (3 * 8)) & 0xFF)
	_ = renderer.SetDrawColor(r, g, b, a) // @FIXME(njs): Handle error
}

func target_rect(t target) sdl.Rect {
	return make_rect(t.x, t.y, TARGET_WIDTH, TARGET_HEIGHT)
}

func proj_rect(x float32, y float32) sdl.Rect {
	return make_rect(x, y, PROJ_SIZE, PROJ_SIZE)
}

func bar_rect(x float32) sdl.Rect {
	return make_rect(x, BAR_Y-BAR_THICCNESS/2, BAR_LEN, BAR_THICCNESS)
}

func horz_collision(dt float32) {
	var proj_nx = float32(proj_x + proj_dx*PROJ_SPEED*dt)
	br := bar_rect(bar_x)
	pr := proj_rect(proj_nx, proj_y)
	if proj_nx < 0 || proj_nx+PROJ_SIZE > WINDOW_WIDTH || pr.HasIntersection(&br) {
		proj_dx *= -1
		return
	}

	for i, _ := range targets_pool {
		pr := proj_rect(proj_nx, proj_y)
		tr := target_rect(targets_pool[i])
		if !targets_pool[i].dead && pr.HasIntersection(&tr) {
			targets_pool[i].dead = true
			proj_dx *= -1
			return
		}
	}
	proj_x = proj_nx
}

func vert_collision(dt float32) {
	proj_ny := float32(proj_y + proj_dy*PROJ_SPEED*dt)
	if proj_ny < 0 || proj_ny+PROJ_SIZE > WINDOW_HEIGHT {
		proj_dy *= -1
		return
	}
	pr := proj_rect(proj_x, proj_ny)
	br := bar_rect(bar_x)
	if pr.HasIntersection(&br) {
		if bar_dx != 0 {
			proj_dx = bar_dx
		}
		proj_dy *= -1
		return
	}
	for i, _ := range &targets_pool {
		pr := proj_rect(proj_x, proj_ny)
		tr := target_rect(targets_pool[i])
		if !targets_pool[i].dead && pr.HasIntersection(&tr) {
			targets_pool[i].dead = true
			proj_dy *= -1
			return
		}
	}

	proj_y = proj_ny
}

func clamp(value float32, minimum float32, maximum float32) float32 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func bar_collision(dt float32) {
	var bar_nx = float32(clamp(bar_x+bar_dx*BAR_SPEED*dt, 0, WINDOW_WIDTH-BAR_LEN))
	pr := proj_rect(proj_x, proj_y)
	br := bar_rect(bar_nx)
	if pr.HasIntersection(&br) {
		return
	}
	bar_x = bar_nx
}

func update(dt float32) {
	if !pause && started {
		pr := proj_rect(proj_x, proj_y)
		br := bar_rect(bar_x)
		if pr.HasIntersection(&br) {
			proj_y = BAR_Y - BAR_THICCNESS/2 - PROJ_SIZE - 1.0
			return
		}
		bar_collision(dt)
		horz_collision(dt)
		vert_collision(dt)
	}
}

func render(renderer *sdl.Renderer) {
	set_color(renderer, PROJ_COLOR)
	pr := proj_rect(proj_x, proj_y)
	renderer.FillRect(&pr)

	set_color(renderer, BAR_COLOR)
	br := bar_rect(bar_x)
	renderer.FillRect(&br)

	set_color(renderer, TARGET_COLOR)
	for _, t := range targets_pool {
		if !t.dead {
			tr := target_rect(t)
			renderer.FillRect(&tr)
		}
	}
}

func main() {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		sdl.Log("Unable to initialize SDL: %s", sdl.GetError())
		// return error.SDLInitializationFailed
		return
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Gobreak", 0, 0, WINDOW_WIDTH, WINDOW_HEIGHT, 0)
	if err != nil {
		sdl.Log("Unable to initialize SDL: %s", sdl.GetError())
		// return error.SDLInitializationFailed
		return
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		sdl.Log("Unable to initialize SDL: %s", sdl.GetError())
		// return error.SDLInitializationFailed
		return
	}
	defer renderer.Destroy()

	keyboard := sdl.GetKeyboardState()

	for !quit {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				{
					quit = true
				}
			case *sdl.KeyboardEvent:
				{
					if t.Type == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_SPACE {
						pause = !pause
					}
				}
			}
		}

		bar_dx = 0
		if keyboard[sdl.SCANCODE_A] != 0 || keyboard[sdl.SCANCODE_LEFT] != 0 {
			bar_dx += -1
			if !started {
				started = true
				proj_dx = -1
			}
		}
		if keyboard[sdl.SCANCODE_D] != 0 || keyboard[sdl.SCANCODE_RIGHT] != 0 {
			bar_dx += 1
			if !started {
				started = true
				proj_dx = 1
			}
		}

		update(DELTA_TIME_SEC)

		set_color(renderer, BACKGROUND_COLOR)
		renderer.Clear()

		render(renderer)

		renderer.Present()

		sdl.Delay(1000 / FPS)
	}
}
