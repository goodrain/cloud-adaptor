package ssh

import "testing"

func TestGetOrMakeSSHRSA(t *testing.T) {
	pub, err := GetOrMakeSSHRSA()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(pub)
}
