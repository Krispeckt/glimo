package instructions

// Display describes the layout model used by a container.
// Only DisplayFlex is currently implemented but the enum allows future extensions.
type Display int

const (
	// DisplayFlex enables Flexbox-style layout behavior.
	DisplayFlex Display = iota
)

// FlexDirection defines the orientation of the main axis in the flex container.
type FlexDirection int

const (
	// Row lays out items horizontally, left-to-right by default.
	Row FlexDirection = iota
	// Column lays out items vertically, top-to-bottom by default.
	Column
)

// JustifyContent controls how free space is distributed along the main axis.
type JustifyContent int

const (
	JustifyStart        JustifyContent = iota // Items packed at the start (default)
	JustifyCenter                             // Items centered along main axis
	JustifyEnd                                // Items packed at the end
	JustifySpaceBetween                       // Even spacing between items, none at ends
	JustifySpaceAround                        // Equal spacing around items, half-space at edges
	JustifySpaceEvenly                        // Equal spacing including container edges
)

// AlignItems controls cross-axis alignment of items within each line.
type AlignItems int

const (
	AlignItemsStart   AlignItems = iota // Align items to the start of the cross axis
	AlignItemsCenter                    // Align items to the cross-axis center
	AlignItemsEnd                       // Align items to the end of the cross axis
	AlignItemsStretch                   // Stretch items to fill the line’s cross size
)

// PositionType indicates whether an item participates in normal layout flow.
type PositionType int

const (
	// PosRelative participates in normal flow (default).
	PosRelative PositionType = iota
	// PosAbsolute is removed from flow and positioned relative to the container padding box.
	PosAbsolute
)

// ContainerStyle defines CSS-like layout properties for an AutoLayout container.
// All numeric units are pixels. Width/Height of 0 mean "auto-size by content".
type ContainerStyle struct {
	Display       Display
	Direction     FlexDirection
	Wrap          bool
	Padding       [4]int  // top, right, bottom, left
	Gap           Vector2 // gap.X = horizontal spacing, gap.Y = vertical spacing
	Justify       JustifyContent
	AlignItems    AlignItems
	AlignContent  AlignItems // cross-axis packing across multiple lines: Start/Center/End/Stretch
	Width, Height int        // container outer dimensions; 0 = auto by content
}

// ItemStyle defines layout behavior of a single child within a flex container.
type ItemStyle struct {
	Margin     [4]int // top, right, bottom, left
	Width      int    // fixed width; 0 = auto
	Height     int    // fixed height; 0 = auto
	FlexGrow   float64
	FlexShrink float64 // defaults to 1 if 0 and negative free space exists
	FlexBasis  int     // preferred main size in px; 0 = auto → width/height/intrinsic
	AlignSelf  *AlignItems

	// Positioning properties for absolute items.
	Position PositionType
	Top      *int
	Right    *int
	Bottom   *int
	Left     *int

	// Painting order (higher values drawn later).
	ZIndex int

	// IgnoreGapBefore skips the container gap directly before this item.
	// Affects line construction, wrapping, and final positioning.
	IgnoreGapBefore bool
}
