package provider

var knownProviders = []string{
	"ssh",
	"portxd",
	"cloudflare",
	"frp",
}

func KnownProviders() []string {
	return knownProviders
}

func IsKnown(name string) bool {
	for _, known := range knownProviders {
		if known == name {
			return true
		}
	}
	return false
}
