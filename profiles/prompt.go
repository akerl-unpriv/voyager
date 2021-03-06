package profiles

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

// PromptStore is a storage backend which asks the user for input
type PromptStore struct{}

// Lookup asks the user for credentials
func (p *PromptStore) Lookup(profile string) (credentials.Value, error) {
	logger.InfoMsgf("looking up %s in prompt store", profile)
	fmt.Printf("Please enter your credentials for profile: %s\n", profile)
	accessKey, err := p.getUserInput("AWS Access Key: ")
	if err != nil {
		return credentials.Value{}, err
	}
	secretKey, err := p.getUserInput("AWS Secret Key: ")
	if err != nil {
		return credentials.Value{}, err
	}
	return credentials.Value{
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
	}, nil
}

// Check is always false, because user input is never cached
func (p *PromptStore) Check(_ string) bool {
	return false
}

// Delete is a no-op, as Prompt never stores credentials
func (p *PromptStore) Delete(_ string) error {
	return nil
}

func (p *PromptStore) getUserInput(message string) (string, error) {
	infoReader := bufio.NewReader(os.Stdin)
	fmt.Fprint(os.Stderr, message)
	info, err := infoReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	info = strings.TrimSpace(info)
	if info == "" {
		return "", fmt.Errorf("no input provided")
	}
	return info, nil
}
