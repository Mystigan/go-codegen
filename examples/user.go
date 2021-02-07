package examples

import (
	"database/sql"
	"net"
	"time"

	"github.com/google/uuid"
)

//go:generate go run ../main.go -type=User -pkg=model -out=./model/user.go
type User struct {
	ID                  uuid.UUID
	Email               sql.NullString
	EmailVerified       bool
	PhoneNumber         string
	PhoneNumberVerified bool
	Name                string
	FamilyName          string
	GivenName           string
	MiddleName          string
	Nickname            string
	PreferredUsername   string
	// URL of the End-User's profile page
	Profile string
	Picture string
	// URL of the End-User's Web page or blog
	Website string
	// End-User's gender (m|f|o|x)
	Gender string
	// End-User's birthday, represented as an ISO 8601:2004 [ISO8601?2004] YYYY-MM-DD format. The year MAY be 0000, indicating that it is omitted
	Birthdate sql.NullTime
	// String from zoneinfo [zoneinfo] time zone database representing the End-User's time zone. For example, Europe/Paris or America/Los_Angeles
	Zoneinfo string
	// End-User's locale, represented as a BCP47 [RFC5646] language tag. This is typically an ISO 639-1 Alpha-2 [ISO639?1] language code in lowercase and an ISO 3166-1 Alpha-2 [ISO3166?1] country code in uppercase, separated by a dash. For example, en-US or fr-CA. As a compatibility note, some implementations have used an underscore as the separator rather than a dash, for example, en_US
	Locale string
	// Full street address component, which MAY include house number, street name, Post Office Box, and multi-line extended street address information. This field MAY contain multiple lines, separated by newlines. Newlines can be represented either as a carriage return/line feed pair ("\r\n") or as a single line feed character ("\n").
	StreetAddress string
	// City or locality component.
	Locality string
	// State, province, prefecture or region component.
	Region string
	// Zip code or postal code component.
	PostalCode string
	// Country name component.
	Country                string
	ConfirmationToken      sql.NullString
	ConfirmationSentAt     sql.NullTime
	ConfirmedAt            sql.NullTime
	UnconfirmedEmail       string
	ResetPasswordToken     sql.NullString
	ResetPasswordSentAt    sql.NullTime
	AllowPasswordChange    bool
	SignInCount            int32
	CurrentSignInAt        sql.NullTime
	CurrentSignInIp        net.IP
	CurrentSignInUserAgent string
	LastSignInAt           sql.NullTime
	LastSignInIp           net.IP
	LastSignInUserAgent    string
	LastSignOutAt          sql.NullTime
	LastSignOutIp          net.IP
	LastSignOutUserAgent   string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	DeletedAt              sql.NullTime
}
