package xtls

type Cipher interface {
	Encrypt(src string) (string, error)
	Decrypt(src string) (string, error)
}

var cipheres = map[string]Cipher{
	"": &DefaultCipher{},
}

type DefaultCipher struct{}

//Encrypt is method used for encryption
func (c *DefaultCipher) Encrypt(src string) (string, error) {
	return src, nil
}

//Decrypt is method used for decryption
func (c *DefaultCipher) Decrypt(src string) (string, error) {
	return src, nil
}
