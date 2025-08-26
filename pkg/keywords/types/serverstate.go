package types

// ServerState represent game-server state for Arma3
type ServerState byte

const (
	ServerState0 ServerState = iota // no server
	ServerState1                    // server created, no mission selected
	ServerState2                    // mission is in editing phase
	ServerState3                    // mission is selected, assigning roles
	ServerState4                    // mission is in sending phase
	ServerState5                    // game (island, vehicles etc.) is loading
	ServerState6                    // prepared to launch game
	ServerState7                    // game is launched
	ServerState8                    // game is finished
	ServerState9                    // game is aborted
)

// String return string represent of uint32 value in s* GameTags in A2S_INFO for Arma3
func (ss ServerState) String() string {
	switch ss {
	case ServerState0:
		return "NONE"
	case ServerState1:
		return "SELECTING MISSION"
	case ServerState2:
		return "EDITING MISSION"
	case ServerState3:
		return "ASSIGNING ROLES"
	case ServerState4:
		return "SENDING MISSION"
	case ServerState5:
		return "LOADING GAME"
	case ServerState6:
		return "BRIEFING"
	case ServerState7:
		return "PLAYING"
	case ServerState8:
		return "DEBRIEFING"
	case ServerState9:
		return "MISSION ABORTED"
	default:
		return "NONE"
	}
}
