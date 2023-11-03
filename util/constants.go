package util

const (
	// config
	AskForConfig string = "camloc/config"
	GetConfig    string = "camloc/+/config"
	SetConfig    string = "camloc/+/config/set"
	// locate
	GetLocation string = "camloc/+/locate"
	Flash       string = "camloc/+/flash"

	// camera state
	AskForState string = "camloc/state"
	GetState    string = "camloc/+/state"
	SetState    string = "camloc/+/state/set"
	SetAllState string = "camloc/state/set"

	// last will
	Disconnect string = "camloc/+/dc"
)
