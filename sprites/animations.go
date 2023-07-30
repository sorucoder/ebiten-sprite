package sprites

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Direction a possible animation direction (forward/reverse).
type Direction string

// Frame represents a single animation frame.
type Frame struct {
	// Image is the frame's image.
	Image *ebiten.Image

	// Duration is the duration the frame is displayed before progressing to the next frame.
	Duration time.Duration
}

// Animation objects contain the actual images and metadata used by sprites. Sprites are
// responsible for maintaining their own state, such as which frame of an animation is
// currently displayed.
type Animation struct {
	// Frames is a slice of this animation's individual frames.
	Frames []*Frame

	// Direction is the direction the animation will be played (forward/reverse).
	Direction Direction
}
