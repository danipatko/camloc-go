package util

const (
	// config
	AskForConfig string = "camloc/config"
	GetConfig    string = "camloc/+/config/get"
	SetConfig    string = "camloc/+/config/set"
	// locate
	GetLocation string = "camloc/+/locate"
	Flash       string = "camloc/+/flash"

	ForceCameraOn      string = "camloc/camstate/on"
	ForceCameraOff     string = "camloc/camstate/off"
	ForceThisCameraOn  string = "camloc/+/camstate/on"
	ForceThisCameraOff string = "camloc/+/camstate/off"

	// last will
	Disconnect string = "camloc/+/dc"
)
