package sprites

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type (
	HorizonalAlignment func(x float64, width float64) float64
	VerticalAlignment  func(y float64, height float64) float64
	Alignment          func(x float64, y float64, width float64, height float64) (float64, float64)
)

var (
	Left HorizonalAlignment = func(x float64, width float64) float64 {
		return x
	}

	Center HorizonalAlignment = func(x float64, width float64) float64 {
		return x - (width / 2)
	}

	Right HorizonalAlignment = func(x float64, width float64) float64 {
		return x - width
	}

	Top VerticalAlignment = func(y float64, height float64) float64 {
		return y
	}

	Middle VerticalAlignment = func(y float64, height float64) float64 {
		return y - (height / 2)
	}

	Bottom VerticalAlignment = func(y float64, height float64) float64 {
		return y - height
	}

	TopLeft Alignment = func(x float64, y float64, width float64, height float64) (float64, float64) {
		return Left(x, width), Top(y, height)
	}

	TopCenter Alignment = func(x float64, y float64, width float64, height float64) (float64, float64) {
		return Center(x, width), Top(y, height)
	}

	TopRight Alignment = func(x float64, y float64, width float64, height float64) (float64, float64) {
		return Right(x, width), Top(y, height)
	}

	MiddleLeft Alignment = func(x float64, y float64, width float64, height float64) (float64, float64) {
		return Left(x, width), Middle(y, height)
	}

	MiddleCenter Alignment = func(x float64, y float64, width float64, height float64) (float64, float64) {
		return Center(x, width), Middle(y, height)
	}

	MiddleRight Alignment = func(x float64, y float64, width float64, height float64) (float64, float64) {
		return Right(x, width), Middle(y, height)
	}

	BottomLeft Alignment = func(x float64, y float64, width float64, height float64) (float64, float64) {
		return Left(x, width), Bottom(y, height)
	}

	BottomMiddle Alignment = func(x float64, y float64, width float64, height float64) (float64, float64) {
		return Middle(x, width), Bottom(y, height)
	}

	BottomRight Alignment = func(x float64, y float64, width float64, height float64) (float64, float64) {
		return Right(x, width), Bottom(y, height)
	}
)

type Sprite struct {
	X      float64
	Y      float64
	Origin Alignment
	Scale  float64
	Speed  float64
	Angle  float64

	animation *Animation
	frame     int
	paused    bool
	Visible   bool
	repeat    bool
	last      time.Time
	options   *ebiten.DrawImageOptions
}

func NewSprite(animation *Animation) *Sprite {
	return &Sprite{
		X:       0.0,
		Y:       0.0,
		Origin:  TopLeft,
		Scale:   1.0,
		Speed:   1.0,
		Angle:   0.0,
		Visible: true,

		animation: animation,
		frame:     0,
		paused:    false,
		repeat:    true,
		last:      time.Now(),
		options:   new(ebiten.DrawImageOptions),
	}
}

func (sprite *Sprite) SetAnimation(animation *Animation, repeat bool) {
	sprite.animation = animation
	sprite.frame = 0
	sprite.paused = true
	sprite.repeat = repeat
	sprite.last = time.Now()
}

func (sprite *Sprite) Start(at time.Time) {
	if !sprite.paused {
		return
	}

	if at.IsZero() {
		at = time.Now()
	}

	sprite.frame = 0
	sprite.last = at
	sprite.paused = false
}

func (sprite *Sprite) Stop() {
	if sprite.paused {
		return
	}

	sprite.frame = 0
	sprite.last = time.Now()
	sprite.paused = true
}

func (sprite *Sprite) Update(at time.Time) {
	if sprite.paused {
		return
	}

	elapsed := at.Sub(sprite.last)
	duration := sprite.animation.Frames[sprite.frame].Duration * time.Duration(sprite.Speed)
	if elapsed < duration {
		return
	}

	if sprite.frame >= len(sprite.animation.Frames)-1 {
		if sprite.repeat {
			sprite.frame = 0
		} else {
			sprite.paused = true
		}
	} else {
		sprite.frame++
	}
	sprite.last = at
}

func (sprite *Sprite) Draw(target *ebiten.Image) {
	if !sprite.Visible {
		return
	}

	frame := sprite.animation.Frames[sprite.frame]
	bounds := frame.Image.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	drawWidth := float64(width) * sprite.Scale
	drawHeight := float64(height) * sprite.Scale
	drawX, drawY := sprite.Origin(sprite.X, sprite.Y, drawWidth, drawHeight)

	sprite.options.GeoM.Reset()
	sprite.options.GeoM.Scale(sprite.Scale, sprite.Scale)
	sprite.options.GeoM.Translate(drawX, drawY)
	sprite.options.GeoM.Rotate(sprite.Angle)

	target.DrawImage(frame.Image, sprite.options)
}
