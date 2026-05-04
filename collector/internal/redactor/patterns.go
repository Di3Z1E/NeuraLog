package redactor

var builtinPatterns = []struct {
	name    string
	pattern string
	replace string
}{
	{"JWT", `eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`, "[REDACTED:JWT]"},
	{"BearerToken", `(?i)bearer\s+[A-Za-z0-9\-._~+/]+=*`, "[REDACTED:BEARER_TOKEN]"},
	{"AWSKeyID", `(?:AKIA|ASIA|AROA)[A-Z0-9]{16}`, "[REDACTED:AWS_KEY_ID]"},
	{"AWSSecretKey", `(?i)aws.{0,20}secret.{0,20}[=:]\s*[A-Za-z0-9/+]{40}`, "[REDACTED:AWS_SECRET]"},
	{"GenericAPIKey", `(?i)(?:api[_-]?key|apikey|x-api-key)\s*[=:]\s*[^\s"'&]{16,}`, "[REDACTED:API_KEY]"},
	{"Password", `(?i)(?:password|passwd|pwd)\s*[=:]\s*[^\s"'&]{6,}`, "[REDACTED:PASSWORD]"},
	{"GenericSecret", `(?i)(?:secret|token)\s*[=:]\s*[^\s"'&]{16,}`, "[REDACTED:SECRET]"},
	{"DatabaseURL", `(?i)(?:postgres|postgresql|mysql|mongodb|redis|amqp)://[^\s"']+:[^\s"']+@[^\s"']+`, "[REDACTED:DB_URL]"},
	{"BasicAuthURL", `(https?://)[^:@\s"']+:[^@\s"']+@`, "${1}[REDACTED:CREDENTIALS]@"},
	{"PrivateKeyBlock", `-----BEGIN (?:[A-Z]+ )?PRIVATE KEY-----[\s\S]{0,4096}?-----END (?:[A-Z]+ )?PRIVATE KEY-----`, "[REDACTED:PRIVATE_KEY]"},
	{"CreditCard", `\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13})\b`, "[REDACTED:CARD]"},
}
