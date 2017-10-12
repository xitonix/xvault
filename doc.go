// Package xvault provides the functionality of encrypting/decrypting any io.Reader into one or more io.Writer outputs.
// You can manually use the provided Encoder and Decoder types or automate the encryption/decryption tasks by implementing
// a new Vault and passing it to a VaultProcessor object. Check FilesystemVault to see an example.
package xvault
