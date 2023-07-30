package asesprite

import (
	unmarshaller "encoding/json"
	"errors"
	"fmt"
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sorucoder/ebiten-sprite/sprites"
)

// Direction is the direction an animation should be played.
type direction string

// BlendMode is the Asesprite blend mode of a layer.
type blendMode string

type rectangle struct {
	coordinates
	dimensions
}

type coordinates struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type dimensions struct {
	Width  int `json:"w"`
	Height int `json:"h"`
}

type frameTag struct {
	Name      string    `json:"name"`
	From      int       `json:"from"`
	To        int       `json:"to"`
	Direction direction `json:"direction"`
}

// layer represents an Asesprite layer.
type layer struct {
	Name      string    `json:"name"`
	Opacity   uint8     `json:"opacity"`
	BlendMode blendMode `json:"blendMode"`
}

// slice represents an Asesprite slice.
type slice struct {
}

// metadata represents an Asesprite sprite sheet's metadata.
type metadata struct {
	Application string     `json:"app"`
	Version     string     `json:"version"`
	Image       string     `json:"image"`
	Format      string     `json:"format"`
	Size        dimensions `json:"size"`
	Scale       string     `json:"scale"`
	FrameTags   []frameTag `json:"frameTags"`
	Layers      []layer    `json:"layers"`
	Slices      []slice    `json:"slice"`
}

// frame represents a single Asesprite frame in the array-style sprite sheet format.
type frame struct {
	Filename         string        `json:"filename"`
	Frame            rectangle     `json:"frame"`
	Rotated          bool          `json:"rotated"`
	Trimmed          bool          `json:"trimmed"`
	SpriteSourceSize rectangle     `json:"spriteSourceSize"`
	SourceSize       dimensions    `json:"sourceSize"`
	Duration         time.Duration `json:"duration"`
}

// asespriteData is the top-level entity of an Asesprite sprite sheet data file.  Currently,
// only sprite sheets exported in the array-style format are supported.
type asespriteData struct {
	Frames   []frame  `json:"frames"`
	Metadata metadata `json:"meta"`
}

// asespriteSpriteSheet is the logical grouping of an Asesprite sprite sheet's image and its
// accompanying data.
type asespriteSpriteSheet struct {
	Data           *asespriteData
	Image          *ebiten.Image
	imageCache     []*ebiten.Image
	animationCache map[string]*sprites.Animation
}

func (s *asespriteSpriteSheet) Animation(name string) (*sprites.Animation, error) {
	// First, check the animation animationCache.
	animation, ok := s.animationCache[name]
	if ok {
		return animation, nil
	}

	// Find the frame tag with the matching name.
	var tag *frameTag

	for _, t := range s.Data.Metadata.FrameTags {
		if t.Name == name {
			tag = &t
			break
		}
	}

	// Frame tag not found.
	if tag == nil {
		return nil, AnimationNotFoundError(name)
	}

	// Create a slice of the frames.
	frames := make([]*sprites.Frame, 0, tag.To-tag.From+1)

	for i := tag.From; i <= tag.To; i++ {
		f := s.Data.Frames[i]

		// Check image cache for existing image.
		img := s.imageCache[i]

		// Cache miss.
		if img == nil {

			img, ok = s.Image.SubImage(image.Rect(
				f.Frame.X,
				f.Frame.Y,
				f.Frame.X+f.Frame.Width,
				f.Frame.Y+f.Frame.Height)).(*ebiten.Image)

			// As of Ebitengine 2.3.3, SubImage always returns *ebiten.Image.  This check
			// is in place in case of future changes to this behavior, as well as changes
			// to the image.Image interface.
			if !ok {
				return nil, errors.New("failed to cast image.Image to ebiten.Image")
			}

			// Add image to cache.
			s.imageCache[i] = img
		}

		frames = append(frames, &sprites.Frame{Image: img, Duration: time.Millisecond * f.Duration})
	}
	// Create the animation and add it to the animationCache.
	animation = &sprites.Animation{Frames: frames, Direction: sprites.Direction(tag.Direction)}
	s.animationCache[name] = animation

	return animation, nil
}

func (s *asespriteSpriteSheet) AllAnimations() (map[string]*sprites.Animation, error) {
	animations := make(map[string]*sprites.Animation)

	for _, tag := range s.Data.Metadata.FrameTags {
		a, err := s.Animation(tag.Name)

		if err != nil {
			return nil, err
		}

		animations[tag.Name] = a
	}

	return animations, nil
}

// NewSpriteSheetLoader returns an implementation of SpriteSheetLoader from an Asesprite JSON
// payload and a sprite sheet image.  Sprite sheet data must be in array-style format; hash
// format is unsupported at this time.
func NewSpriteSheetLoader(json []byte, image image.Image) (sprites.SpriteSheetLoader, error) {
	// Unmarshall JSON data.
	data := &asespriteData{}

	if err := unmarshaller.Unmarshal(json, data); err != nil {
		return nil, fmt.Errorf("unable to unmarshall Asesprite JSON data: %w", err)
	}

	return &asespriteSpriteSheet{
		data, ebiten.NewImageFromImage(image),
		make([]*ebiten.Image, len(data.Frames)),
		make(map[string]*sprites.Animation),
	}, nil
}
