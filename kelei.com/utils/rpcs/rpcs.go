package rpcs

type Args struct {
	V map[string]interface{}
}

type Reply struct {
	RS *string
	SC string
}
