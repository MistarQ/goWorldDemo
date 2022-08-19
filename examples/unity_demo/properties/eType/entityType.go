package eType

// type
const (
	Account   = "Account"
	Player    = "Player"
	Monster   = "Monster"
	BlackHole = "BlackHole"
)

func IsPlayer(typeName string) bool {
	switch typeName {
	case Player:
		return true
	}
	return false
}

func IsMonster(typeName string) bool {
	switch typeName {
	case Monster:
		return true
	}
	return false
}
