package constant

const (
	PublicIDPrefixNote = "nte_"

	NoteColorDefault = "default"
	NoteColorCoral   = "coral"
	NoteColorSage    = "sage"
	NoteColorSky     = "sky"
	NoteColorBlush   = "blush"
	NoteColorSand    = "sand"
	NoteColorLilac   = "lilac"
)

// ValidNoteColors is the allowed set for notes.color.
var ValidNoteColors = map[string]struct{}{
	NoteColorDefault: {},
	NoteColorCoral:   {},
	NoteColorSage:    {},
	NoteColorSky:     {},
	NoteColorBlush:   {},
	NoteColorSand:    {},
	NoteColorLilac:   {},
}
