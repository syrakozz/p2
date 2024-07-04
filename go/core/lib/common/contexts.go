package common

type (
	timeoutKeyType      string
	meKeyType           string
	accountKeyType      string
	firebaseUserKeyType string
)

const (
	// TimeoutKey is a context value.
	TimeoutKey timeoutKeyType = "timeout"

	//MeKey is a context value for current logged in user Firebase UserRecord.
	MeKey meKeyType = "me"

	//AccountKey is a context value for current logged in Firestore account.
	AccountKey accountKeyType = "account"

	//FirebaseUserKey is a context value for current logged in Firestore account.
	FirebaseUserKey firebaseUserKeyType = "fbuser"
)
