package discord

type MemberDefinition struct {
	Nickname  string
	FirstName string
	LastName  string
	ID        string
}

func (m *MemberDefinition) Mention() string {
	return Mention(m.ID)
}

// Friends contains static map of discord id to a member definition
var Friends = map[string]MemberDefinition{
	"611471485218848768": {
		Nickname:  "Amaterasu",
		FirstName: "Paulina",
		LastName:  "Grzanek",
		ID:        "611471485218848768",
	},
	"144797044236615680": {
		Nickname:  "Mawgan",
		FirstName: "Artur",
		LastName:  "Wieczorek",
		ID:        "144797044236615680",
	},
	"300692223769575425": {
		Nickname:  "Ravutto",
		FirstName: "Rafał",
		LastName:  "Czarnecki",
		ID:        "300692223769575425",
	},
	"498528474408157232": {
		Nickname:  "Tamako",
		FirstName: "Tamako",
		LastName:  "Lumisade",
		ID:        "498528474408157232",
	},
	"600656821568405508": {
		Nickname:  "Walter_441",
		FirstName: "Zachary",
		LastName:  "Bujok",
		ID:        "600656821568405508",
	},
	"691271489907064922": {
		Nickname:  "PureGold",
		FirstName: "Wojtek",
		LastName:  "Górny",
		ID:        "691271489907064922",
	},
	"417397822372052992": {
		Nickname:  "Szat265",
		FirstName: "Emila",
		LastName:  "Gerlicka",
		ID:        "417397822372052992",
	},
	"170268157494296576": {
		Nickname:  "Taliön",
		FirstName: "Przemek",
		LastName:  "Żydek",
		ID:        "170268157494296576",
	},
}

func GetFriend(id string) (*MemberDefinition, bool) {
	friend, ok := Friends[id]

	return &friend, ok
}
