package driver

import (
	"context"
	"testing"
)

func TestGetDriverCreateOptionsAppliesSecretDefaults(t *testing.T) {
	old := defaultCredentialLoader
	defaultCredentialLoader = func(_ context.Context) credentialDefaults {
		return credentialDefaults{Org: "preset-org", RefreshToken: "preset-token"}
	}
	defer func() { defaultCredentialLoader = old }()

	d := NewDriver()
	flags, err := d.GetDriverCreateOptions(context.Background())
	if err != nil {
		t.Fatalf("GetDriverCreateOptions() error = %v", err)
	}

	orgFlag := flags.Options[flagOrganization]
	if orgFlag.Default == nil || orgFlag.Default.DefaultString != "preset-org" {
		t.Fatalf("org Default = %v, want DefaultString=preset-org", orgFlag.Default)
	}
	tokenFlag := flags.Options[flagRefreshToken]
	if tokenFlag.Default == nil || tokenFlag.Default.DefaultString != "preset-token" {
		t.Fatalf("token Default = %v, want DefaultString=preset-token", tokenFlag.Default)
	}
}

func TestGetDriverCreateOptionsNoDefaultsWhenNoSecret(t *testing.T) {
	old := defaultCredentialLoader
	defaultCredentialLoader = func(_ context.Context) credentialDefaults {
		return credentialDefaults{}
	}
	defer func() { defaultCredentialLoader = old }()

	d := NewDriver()
	flags, err := d.GetDriverCreateOptions(context.Background())
	if err != nil {
		t.Fatalf("GetDriverCreateOptions() error = %v", err)
	}

	orgFlag := flags.Options[flagOrganization]
	if orgFlag.Default != nil {
		t.Fatalf("org Default = %v, want nil when no secret", orgFlag.Default)
	}
}
