package asesprite

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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

// AsespriteSpriteSheet is the logical grouping of an Asesprite sprite sheet's image and its
// accompanying data.
type AsespriteSpriteSheet struct {
	data           *asespriteData
	image          *ebiten.Image
	imageCache     []*ebiten.Image
	animationCache map[string]*sprites.Animation
}

func (spritesheet *AsespriteSpriteSheet) Animation(name string) (*sprites.Animation, error) {
	// First, check the animation animationCache.
	animation, ok := spritesheet.animationCache[name]
	if ok {
		return animation, nil
	}

	// Find the frame tag with the matching name.
	var tag *frameTag

	for _, t := range spritesheet.data.Metadata.FrameTags {
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
		f := spritesheet.data.Frames[i]

		// Check image cache for existing image.
		img := spritesheet.imageCache[i]

		// Cache miss.
		if img == nil {

			img, ok = spritesheet.image.SubImage(image.Rect(
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
			spritesheet.imageCache[i] = img
		}

		frames = append(frames, &sprites.Frame{Image: img, Duration: time.Millisecond * f.Duration})
	}
	// Create the animation and add it to the animationCache.
	animation = &sprites.Animation{Frames: frames, Direction: sprites.Direction(tag.Direction)}
	spritesheet.animationCache[name] = animation

	return animation, nil
}

func (s *AsespriteSpriteSheet) AllAnimations() (map[string]*sprites.Animation, error) {
	animations := make(map[string]*sprites.Animation)

	for _, tag := range s.data.Metadata.FrameTags {
		a, err := s.Animation(tag.Name)

		if err != nil {
			return nil, err
		}

		animations[tag.Name] = a
	}

	return animations, nil
}

// NewSpritesheet returns an implementation of SpriteSheetLoader from an Asesprite JSON payload and a sprite sheet image.
// Sprite sheet data must be in array-style format; hash format is unsupported at this time.
func NewSpritesheet(decodedImage image.Image, jsonBytes []byte) (sprites.Spritesheet, error) {
	var jsonData asespriteData
	if errUnmarshal := json.Unmarshal(jsonBytes, &jsonData); errUnmarshal != nil {
		return nil, fmt.Errorf("failed to unmarshal Asesprite JSON data: %w", errUnmarshal)
	}

	spritesheet := &AsespriteSpriteSheet{
		data:           &jsonData,
		image:          ebiten.NewImageFromImage(decodedImage),
		imageCache:     make([]*ebiten.Image, len(jsonData.Frames)),
		animationCache: make(map[string]*sprites.Animation),
	}

	return spritesheet, nil
}

// NewSpritesheet returns an implementation of SpriteSheetLoader from disk.
// Sprite sheet data must be in array-style format; hash format is unsupported at this time.
func NewSpritesheetFromFiles(imagePath string, jsonPath string) (sprites.Spritesheet, error) {
	spritesheetImage, _, errOpenImage := ebitenutil.NewImageFromFile(imagePath)
	if errOpenImage != nil {
		return nil, fmt.Errorf(`failed to open image file: %w`, errOpenImage)
	}

	jsonFile, errReadJSON := os.Open(jsonPath)
	if errReadJSON != nil {
		return nil, fmt.Errorf(`failed to read Asesprite JSON file: %w`, errOpenImage)
	}
	defer jsonFile.Close()
	var spritesheetData asespriteData
	decoder := json.NewDecoder(jsonFile)
	if errDecode := decoder.Decode(&spritesheetData); errDecode != nil {
		return nil, fmt.Errorf("failed to decode Asesprite JSON data: %w", errDecode)
	}

	spritesheet := &AsespriteSpriteSheet{
		data:           &spritesheetData,
		image:          spritesheetImage,
		imageCache:     make([]*ebiten.Image, len(spritesheetData.Frames)),
		animationCache: make(map[string]*sprites.Animation),
	}

	return spritesheet, nil
}

// NewSpritesheet returns an implementation of SpriteSheetLoader from a filesystem.
// Sprite sheet data must be in array-style format; hash format is unsupported at this time.
func NewSpritesheetFromFileSystem(filesystem fs.FS, imagePath string, jsonPath string) (sprites.Spritesheet, error) {
	spritesheetImage, _, errOpenImage := ebitenutil.NewImageFromFileSystem(filesystem, imagePath)
	if errOpenImage != nil {
		return nil, fmt.Errorf(`failed to open image file: %w`, errOpenImage)
	}

	jsonFile, errReadJSON := filesystem.Open(jsonPath)
	if errReadJSON != nil {
		return nil, fmt.Errorf(`failed to read Asesprite JSON file: %w`, errOpenImage)
	}
	defer jsonFile.Close()
	var spritesheetData asespriteData
	decoder := json.NewDecoder(jsonFile)
	if errDecode := decoder.Decode(&spritesheetData); errDecode != nil {
		return nil, fmt.Errorf("failed to decode Asesprite JSON data: %w", errDecode)
	}

	spritesheet := &AsespriteSpriteSheet{
		data:           &spritesheetData,
		image:          spritesheetImage,
		imageCache:     make([]*ebiten.Image, len(spritesheetData.Frames)),
		animationCache: make(map[string]*sprites.Animation),
	}

	return spritesheet, nil
}
