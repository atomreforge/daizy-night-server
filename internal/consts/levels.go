package consts

type LevelNotificationEmergency string
type LevelNotificationImportance int
type LevelNotificationAddtional string
type Levels struct {
	Emergence  LevelNotificationEmergency
	Importance LevelNotificationImportance
	Addtional  LevelNotificationAddtional
}

// msg that levels A will trigger client intime action.
const (
	A LevelNotificationEmergency = "A"
	B LevelNotificationEmergency = "B"
	C LevelNotificationEmergency = "C"
	D LevelNotificationEmergency = "D"
)

// msg that levels >= 8 will trigger client intime action.
const (
	Zero LevelNotificationImportance = iota
	One
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
)

const ()
