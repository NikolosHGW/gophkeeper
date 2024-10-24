package entity

type TokenHolder struct {
	Token string
}

type LoginPasswordData struct {
	Login    string
	Password string
	URL      string
}

type TextData struct {
	Text string
}

type BinaryData struct {
	FileName    string
	FileContent []byte
}

type BankCardData struct {
	CardNumber string
	ExpiryDate string
	CVV        string
	HolderName string
}
