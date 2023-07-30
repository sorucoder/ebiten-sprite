package sprites

// SpriteSheetLoader loads named Animations from a sprite sheet.
type SpriteSheetLoader interface {
	// Animation returns the tagged animation with the specified name.
	Animation(name string) (*Animation, error)

	// AllAnimations returns a mapping of all tagged animations and their names.
	AllAnimations() (map[string]*Animation, error)
}
