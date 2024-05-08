	identityProviderConfigType := IdentityProviderConfigType
	input.IdentityProviderConfig = &svcsdk.IdentityProviderConfig{
		Name: r.ko.Spec.OIDC.IdentityProviderConfigName,
		Type: &identityProviderConfigType,
	}
