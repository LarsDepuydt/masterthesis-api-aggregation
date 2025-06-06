package data

// HardcodedIDToEmailMapData maps canonical Room IDs (from Building/FMS)
// to their corresponding Outlook email addresses, based on the lists provided.
// This map is used by the entity resolver to quickly find the email
// associated with a Room ID when the Gateway requests Outlook-specific fields.
var HardcodedIDToEmailMapData = map[string]string{
	// Matches from TMV25-Stue section
	"TMV25-Stue-A.006": "tmv25-a.006@adm.aau.dk",
	"TMV25-Stue-C.009": "tmv25-c.009@adm.aau.dk",
	"TMV25-Stue-C.004": "tmv25-c.004@adm.aau.dk",

	// Matches from TMV25-1. Sal section
	"TMV25-1. Sal-A.111":  "tmv25-a.111@adm.aau.dk",
	"TMV25-1. Sal-A.112":  "tmv25-a.112@adm.aau.dk",
	"TMV25-1. Sal-A.115":  "tmv25-a.115@adm.aau.dk",
	"TMV25-1. Sal-A.119a": "tmv25-a.119a@adm.aau.dk",
	"TMV25-1. Sal-A.120a": "tmv25-a.120a@adm.aau.dk",
	"TMV25-1. Sal-B.106a": "tmv25-b.106a@adm.aau.dk",
	"TMV25-1. Sal-C.104a": "tmv25-c.104a@adm.aau.dk",
	"TMV25-1. Sal-C.104b": "tmv25-c.104b@adm.aau.dk",
	"TMV25-1. Sal-C.107":  "tmv25-c.107@adm.aau.dk",
	"TMV25-1. Sal-C.106":  "tmv25-c.106@adm.aau.dk",
	"TMV25-1. Sal-C.102":  "tmv25-c.102@adm.aau.dk",

	// Matches from TMV25-2. Sal section
	"TMV25-2. Sal-A.202":  "tmv25-a.202@adm.aau.dk",
	"TMV25-2. Sal-A.207":  "tmv25-a.207@adm.aau.dk",
	"TMV25-2. Sal-A.212":  "tmv25-a.212@adm.aau.dk",
	"TMV25-2. Sal-A.218a": "tmv25-a.218a@adm.aau.dk",
	"TMV25-2. Sal-A.218b": "tmv25-a.218b@adm.aau.dk",
	"TMV25-2. Sal-B.203":  "tmv25-b.203@adm.aau.dk",
	"TMV25-2. Sal-A.201a": "tmv25-a.201b@adm.aau.dk",
	"TMV25-2. Sal-A.201b": "tmv25-a.201a@adm.aau.dk",

	// Matches from TMV25-3. Sal section
	"TMV25-3. Sal-A.301": "tmv25-a.301@adm.aau.dk",
	"TMV25-3. Sal-A.310": "tmv25-a.310@adm.aau.dk",
	"TMV25-3. Sal-A.313": "tmv25-a.313@adm.aau.dk",
	"TMV25-3. Sal-A.314": "tmv25-a.314@adm.aau.dk",
	"TMV25-3. Sal-A.315": "tmv25-a.315@adm.aau.dk",
}

// HardcodedEmailToIDMapData maps Outlook email addresses to canonical Room IDs.
// This map is populated automatically by the init() function.
// It's the reverse lookup for HardcodedIDToEmailMapData.
var HardcodedEmailToIDMapData map[string]string

// init runs automatically when the package is initialized.
// We use it to populate the HardcodedEmailToIDMapData from HardcodedIDToEmailMapData.
func init() {
	// Initialize the reverse map with a reasonable capacity
	HardcodedEmailToIDMapData = make(map[string]string, len(HardcodedIDToEmailMapData))

	// Populate the reverse map
	for id, email := range HardcodedIDToEmailMapData {
		HardcodedEmailToIDMapData[email] = id
	}
}
