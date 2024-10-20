package mongo

var (
	DefaultConfig = Config{
		URI: "",

		// Database
		DatabaseName: "dalhaega",

		// Collections
		MainCharactersCollection: "main_characters",
		CharactersCollection:     "characters",
	}
)

type Config struct {
	URI string `mapstructure:"mongo_uri"`

	// Database
	DatabaseName string

	// Collections
	MainCharactersCollection string
	CharactersCollection     string
}
